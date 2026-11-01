#!/usr/bin/env python3
"""Independent verifier for upstream candidate fix/blizzard-shard-offsets (d2lod-bot card verifyblizzard).

Written from disasm.py reads of game.exe 1.14d only, before reading the author's tools or docs:
FUN_005a9820 (the shard dropper), FUN_0045c3e0 (seed roll), FUN_00650e40 (seed init),
FUN_0064a3b0 (missile age = data +0xe - data +0x10), FUN_0045adf0 / FUN_0045ae20 -> FUN_006488c0 /
FUN_00648900 (a missile's subtile x / y = path +2 / +6), FUN_0064a280 / FUN_0064a210 (skill / level),
the callers FUN_005aea60 (srvdo 10, Blizzard, mask 5) and FUN_005af880 (srvdo 25, Eruption, mask 0x45).

FUN_005a9820(eax every, esi missile; game, radius, submissile, mask), callee pops 0x10:
  every == 0 or age % every != 0 (signed idiv)  -> nothing
  owner (FUN_00552fd0) or room (FUN_00620bb0) missing -> nothing
  seed {cx + age, 666}; r = radius - 1; dx = roll(2r) - r; dy = roll(2r) - r (x first)
  FUN_0064cb30(room, cx + dx, cy + dy, mask) != 0 (blocked) -> nothing, else FUN_0059fa30 makes it.
roll(n): n < 1 -> 0, seed untouched; else step {lo, hi} = lo * 0x6AC690C5 + hi, then lo & (n-1)
for a power of two, else lo % n (unsigned).

game side: game.exe's own FUN_005a9820 under unicorn (upstream tools/emu/emu.py), native with
FUN_0045c3e0, FUN_00650e40, FUN_0064a3b0, FUN_0045adf0, FUN_0045ae20, FUN_006488c0, FUN_00648900,
FUN_0064a280, FUN_0064a210, FUN_00681ef0 (memset). Hooked (not the claim): FUN_00552fd0 owner
(a fake pointer or 0), FUN_00620bb0 room (a fake pointer or 0), FUN_0064cb30 collision (the case's
blocked rule, every query logged), FUN_0059fa30 missile maker (logs params +0x14 x, +0x18 y, +0x10 id).

    python verify_blizzard.py gen N SEED > cases.tsv
    python verify_blizzard.py game cases.tsv > game.tsv     # run from upstream tools/emu, D2_MPQ_DIR with bin/game.exe
    python verify_blizzard.py model cases.tsv > model.tsv   # this file's own reading of the rule
    python verify_blizzard.py compare cases.tsv game.tsv model.tsv
    python verify_blizzard.py grid cases.tsv game.tsv                # offsets histogram of the radius-7 cases
    python verify_blizzard.py gen-d2d N SEED > d2dcases.txt
    python verify_blizzard.py patch TEST_GAME_CPP           # paste verify_blizzard_mode.inc (throwaway clone only)
    python verify_blizzard.py d2d d2d.tsv                   # histogram of the d2d probe's shard offsets

cases.tsv: id every radius cx cy frame start owner room block mask
  block: none | all | parity (blocked when (x + y) odd) | hashK (blocked when hash(x, y) % K == 0)
game.tsv / model.tsv: id outcome [dx dy] [seedlo seedhi]
  outcome: none (no roll), blocked dx dy, shard dx dy; seed after the call (the missile's +0x20).
"""
import random
import struct
import sys

M32 = 0xFFFFFFFF
MUL = 0x6AC690C5


def gen(n, seed):
    rng = random.Random(seed)
    rows = []

    def add(every, radius, cx, cy, frame, start, owner=1, room=1, block="none", mask=5):
        rows.append((every, radius, cx, cy, frame, start, owner, room, block, mask))

    # edges: exhaustive small grid over the parameters that matter to the rule
    for every in (0, 1, 4, 6, 7, -4):
        for radius in (-3, -1, 0, 1, 2, 3, 5, 7, 8, 9, 17, 100):
            for cx in (0, 1, 6, 32767, 32768, 65529, 65535):
                for age in (0, 1, 3, 4, 8, 12, 96, 100, 32767, -1, -4, -32768):
                    add(every, radius, cx, 1000, age & 0xFFFF if age >= 0 else age, 0)
    # frame/start pairs whose difference crosses int16 (movsx on both)
    for frame, start in ((0, 0), (100, 96), (-32768, 0), (32767, -32768), (-1, 1), (4, -4), (1000, 999)):
        for cx in (0, 65535, 4321):
            add(4, 7, cx, 500, frame, start)
    # owner / room missing, blocked patterns, Eruption's mask
    for owner, room in ((0, 1), (1, 0), (0, 0)):
        for age in (0, 4):
            add(4, 7, 100, 100, age, 0, owner, room)
    for block in ("all", "parity", "hash3", "hash7"):
        for k in range(200):
            add(4 if k % 2 else 6, 7, rng.randrange(65536), rng.randrange(65536), 4 * rng.randrange(800), 0, block=block, mask=5 if k % 2 else 0x45)
    # bulk: Blizzard / Eruption radius 7, every 4 / 6, random spot and age (all ages, so some skip)
    while len(rows) < n:
        k = len(rows)
        every = 4 if k % 2 else 6
        cx, cy = rng.randrange(65536), rng.randrange(65536)
        start = rng.randrange(-32768, 32768)
        frame = start + rng.randrange(0, 200) if k % 5 else rng.randrange(-32768, 32768)
        frame = ((frame + 32768) & 0xFFFF) - 32768
        block = ("none", "none", "none", "parity", "hash5")[k % 5]
        radius = 7 if k % 10 else rng.choice((1, 2, 3, 4, 5, 6, 8, 9, 13, 33))
        add(every, radius, cx, cy, frame, start, block=block, mask=5 if every == 4 else 0x45)
    for i, r in enumerate(rows):
        print(i, *r, sep="\t")


def gen_d2d(n, seed):
    """Cases for the d2d probe: Blizzard and Eruption casts, edge rng states first."""
    rng = random.Random(seed)
    k = 0
    for low, high in ((0, 0), (0, 666), (M32, M32), (1, 0), (0x80000000, 0x7FFFFFFF), (12, 13)):
        for name in ("Blizzard", "Eruption"):
            print(k, name, low, high, 1, 3.0, 0.0)
            k += 1
    while k < n:
        name = "Eruption" if k % 5 == 0 else "Blizzard"
        print(k, name, rng.getrandbits(32), rng.getrandbits(32), rng.randrange(1, 31), round(rng.uniform(-6, 6), 2), round(rng.uniform(-6, 6), 2))
        k += 1


def read_cases(path):
    out = []
    for line in open(path):
        f = line.split()
        out.append((int(f[0]), int(f[1]), int(f[2]), int(f[3]), int(f[4]), int(f[5]), int(f[6]), int(f[7]), int(f[8]), f[9], int(f[10])))
    return out


def blocked(rule, x, y):
    if rule == "none":
        return False
    if rule == "all":
        return True
    if rule == "parity":
        return (x + y) & 1 == 1
    k = int(rule[4:])
    return (((x * 73856093) ^ (y * 19349663)) & M32) % k == 0


def s16(v):
    v &= 0xFFFF
    return v - 0x10000 if v & 0x8000 else v


def game(path):
    sys.path.insert(0, ".")
    from emu import Emu
    e = Emu()
    game_ = e.alloc(0x100)
    unit = e.alloc(0x100)
    data = e.alloc(0x100)
    pth = e.alloc(0x100)
    e.w32(unit, 3)
    e.w32(unit + 0x14, data)
    e.w32(unit + 0x2C, pth)
    cur = {}
    log = []

    def owner(em):
        return 0x31337000 if cur["owner"] else 0

    def room(em):
        return 0x41414000 if cur["room"] else 0

    def coll(em):
        rm, x, y, mask = em.arg(0), em.arg(1), em.arg(2), em.arg(3)
        x, y = x & M32, y & M32
        hit = blocked(cur["block"], x, y)
        log.append(("q", x, y, mask, rm))
        return 1 if hit else 0

    def make(em):
        from unicorn.x86_const import UC_X86_REG_EDX
        p = em.mu.reg_read(UC_X86_REG_EDX)
        log.append(("m", em.r32(p + 0x14), em.r32(p + 0x18), em.r32(p + 0x10), em.r32(p)))
        return 1

    e.hook(0x552FD0, owner, 0)
    e.hook(0x620BB0, room, 1)
    e.hook(0x64CB30, coll, 4)
    e.hook(0x59FA30, make, 0)
    for (cid, every, radius, cx, cy, frame, start, own, rm, block, mask) in read_cases(path):
        cur.update(owner=own, room=rm, block=block)
        log.clear()
        e.mu.mem_write(pth + 2, struct.pack("<H", cx & 0xFFFF))
        e.mu.mem_write(pth + 6, struct.pack("<H", cy & 0xFFFF))
        e.mu.mem_write(data + 0xA, struct.pack("<hh", 64, 20))          # skill, level: copied, unused by the rule
        e.mu.mem_write(data + 0xE, struct.pack("<hh", s16(frame), s16(start)))
        e.w32(unit + 0x20, 0xDEADBEEF)
        e.w32(unit + 0x24, 0xCAFEF00D)
        e.call(0x5A9820, game_, radius & M32, 77, mask, regs={"eax": every & M32, "esi": unit})
        lo, hi = e.r32(unit + 0x20), e.r32(unit + 0x24)
        q = [l for l in log if l[0] == "q"]
        m = [l for l in log if l[0] == "m"]
        bad = ""
        if len(q) > 1 or len(m) > 1 or (m and not q):
            bad = " ODD"
        if q and q[0][3] != mask:
            bad += " MASK"
        if m and (m[0][1:3] != q[0][1:3] or m[0][3] != 77 or m[0][4] != 3):
            bad += " MAKE"
        if not q:
            print(cid, "none", lo, hi, sep="\t")
        else:
            dx, dy = (q[0][1] - cx) & M32, (q[0][2] - cy) & M32
            dx = dx - (1 << 32) if dx & 0x80000000 else dx
            dy = dy - (1 << 32) if dy & 0x80000000 else dy
            print(cid, "shard" if m else "blocked", dx, dy, lo, hi, sep="\t", end="")
            print(bad)


def roll(seed, n):
    if n < 1:
        return 0
    p = seed[0] * MUL + seed[1]
    seed[0], seed[1] = p & M32, (p >> 32) & M32
    return seed[0] & (n - 1) if n & (n - 1) == 0 else seed[0] % n


def model(path):
    for (cid, every, radius, cx, cy, frame, start, own, rm, block, mask) in read_cases(path):
        lo, hi = 0xDEADBEEF, 0xCAFEF00D
        age = s16(frame) - s16(start)
        sev = (every & M32) - (1 << 32) if every & 0x80000000 else every
        # C remainder (truncates toward zero)
        rem = None if sev == 0 else abs(age) % abs(sev) * (1 if age >= 0 else -1)
        if sev == 0 or rem != 0 or not own or not rm:
            print(cid, "none", lo, hi, sep="\t")
            continue
        seed = [(cx + age) & M32, 666]
        r = radius - 1
        dx = roll(seed, 2 * r) - r
        dy = roll(seed, 2 * r) - r
        out = "blocked" if blocked(block, (cx + dx) & M32, (cy + dy) & M32) else "shard"
        print(cid, out, dx, dy, seed[0], seed[1], sep="\t")


def compare(cases, a, b):
    ca = {l.split("\t")[0]: l.rstrip("\n") for l in open(a)}
    cb = {l.split("\t")[0]: l.rstrip("\n") for l in open(b)}
    ids = [l.split("\t")[0] for l in open(cases)]
    diff = [i for i in ids if ca.get(i) != cb.get(i)]
    kinds = {}
    for i in ids:
        k = ca.get(i, "").split("\t")
        kinds[k[1] if len(k) > 1 else "missing"] = kinds.get(k[1] if len(k) > 1 else "missing", 0) + 1
    print(f"cases {len(ids)} differ {len(diff)} outcomes {kinds}")
    for i in diff[:20]:
        print("A", ca.get(i))
        print("B", cb.get(i))


def grid(cases, path):
    rad = {l.split("\t")[0]: int(l.split("\t")[2]) for l in open(cases)}
    hx, hy, n, off = {}, {}, 0, 0
    for l in open(path):
        f = l.split("\t")
        if f[1] == "none" or rad[f[0]] != 7:
            continue
        dx, dy = int(f[2]), int(f[3])
        n += 1
        hx[dx] = hx.get(dx, 0) + 1
        hy[dy] = hy.get(dy, 0) + 1
        off += not (-6 <= dx <= 5 and -6 <= dy <= 5)
    print(f"radius-7 rolls {n}, off the -6..5 grid {off}")
    print("x", sorted(hx.items()))
    print("y", sorted(hy.items()))


def patch(path):
    src = open(path).read()
    anchor = "    character.panel = panel_stats(*data, character.header, character.items, character.stats);\n"
    assert src.count(anchor) == 1, "anchor not found once"
    mode = open(__file__.replace(".py", "_mode.inc")).read()
    open(path, "w").write(src.replace(anchor, anchor + mode))


def d2d(path):
    hx, hy, n, off, per = {}, {}, 0, 0, {}
    for l in open(path):
        f = l.split()
        if len(f) != 5 or f[1] not in ("Blizzard", "Eruption"):
            continue
        cid, frame, dx, dy = f[0], int(f[2]), int(f[3]), int(f[4])
        n += 1
        per[cid] = per.get(cid, 0) + 1
        hx[dx] = hx.get(dx, 0) + 1
        hy[dy] = hy.get(dy, 0) + 1
        off += not (-6 <= dx <= 5 and -6 <= dy <= 5)
    counts = {}
    for v in per.values():
        counts[v] = counts.get(v, 0) + 1
    print(f"shards {n} in {len(per)} casts (shards per cast: {counts}), off the -6..5 grid {off}")
    print("x", sorted(hx.items()))
    print("y", sorted(hy.items()))


if __name__ == "__main__":
    cmd = sys.argv[1]
    if cmd == "gen":
        gen(int(sys.argv[2]), int(sys.argv[3]))
    elif cmd == "gen-d2d":
        gen_d2d(int(sys.argv[2]), int(sys.argv[3]))
    elif cmd == "game":
        game(sys.argv[2])
    elif cmd == "model":
        model(sys.argv[2])
    elif cmd == "compare":
        compare(sys.argv[2], sys.argv[3], sys.argv[4])
    elif cmd == "grid":
        grid(sys.argv[2], sys.argv[3])
    elif cmd == "patch":
        patch(sys.argv[2])
    elif cmd == "d2d":
        d2d(sys.argv[2])
