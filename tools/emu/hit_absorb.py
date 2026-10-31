#!/usr/bin/env python3
"""game.exe 1.14d's Bone Armor / Cyclone Armor absorb per damage record, then the get-hit (d2lod-bot, PR #134 round 2).

The question (CodeRabbit on PR #134, fight.cpp struck_big): which landed blow of a frame loses the part an
absorb pool takes. game.exe runs, for each damage record in landing order, FUN_0057c1e0 (state event 0xb
through FUN_005c0c30 BEFORE the resistance loop and the total +0x4c at 0x57c3a5) and then FUN_0057cee0 (the
mark FUN_0057ce30 and the size test FUN_0057cb00). Bone Armor's event handler FUN_005c6370 takes from the
record's physical field +0x08 only; Cyclone Armor's FUN_005c87a0 from fire +0x10, cold +0x24, lightning +0x1c
(in that order, FUN_005c8780 per field); the pool is a stat of the state's list (skill row +0x54), set back
after each record, the state removed at 0.

The oracle runs FUN_0057c1e0 natively per record with the real dispatcher FUN_005c0c30 walking a one-node
event list on the target (+0x90: event 0xb, the real handler's address, the skill row index) and the real
handler, then FUN_0057cee0 natively, the unit's seed carrying on, as hit_rule.py. Hooked: stat / state getters
(resistances, reductions, absorb stats 0 = the domain: no resistance, no damage reduction, no element absorb),
FUN_0057c060's damage percent (100), the pool's stat list (a dict), the state graphic / removal calls (recorded),
the mode calls (recorded). Skill table, row and stat ids are placeholders the hooks key on (game.exe reads
them from the row: +0x80 state, +0x56 max stat, +0x54 pool stat).

    python hit_absorb.py gen 50000 7 > cases.tsv
    D2_MPQ_DIR=... python hit_absorb.py --exe game.exe game cases.tsv > game.tsv
    g++ -O2 -std=c++2c -I<d2d>/components/{rules,d2s,txt} -I<d2d>/components -o hit_absorb_harness hit_absorb_harness.cpp
    ./hit_absorb_harness --before < cases.tsv > before.tsv    # d2d 5466cd17's attribution, quoted
    ./hit_absorb_harness < cases.tsv > after.tsv              # the patched tree's rules::absorb_blows + rules::struck_big
    python hit_absorb.py compare game.tsv after.tsv cases.tsv

cases.tsv: id kind maxlife pool armor phys fire cold ltng poison hc seedlo seedhi next
(256ths, whole points x 256 as d2d's blows are whole points; armor 0 Bone Armor, 1 Cyclone Armor; poison 0 or
1/256, only a mark; next 1: the following line is another record on the same unit in the same frame, its
pool / seed columns ignored). Output: id gh pool_left damage_left (the frame's get-hit, the pool after it,
the frame's damage after the absorb, 256ths).
"""
import random
import sys

COLS = "id kind maxlife pool armor phys fire cold ltng poison hc seedlo seedhi next".split()
BONE, CYCLONE = 0x5c6370, 0x5c87a0
STATE, POOL_STAT, MAX_STAT, SKILL_INDEX = 0x10, 0xa4, 0xa5, 1


def gen(n, seed):
    rng = random.Random(seed)
    rows = []

    def frame(kind, maxlife, pool, armor, blows):
        sl, sh = rng.getrandbits(32), rng.getrandbits(32)
        for i, (phys, fire, cold, ltng, poison, hc) in enumerate(blows):
            rows.append([len(rows), kind, maxlife, pool, armor, phys << 8, fire << 8, cold << 8, ltng << 8, poison, hc, sl, sh,
                         0 if i == len(blows) - 1 else 1])

    # The review's cases: a physical blow, then a fire blow, Cyclone Armor's pool over the fire; the mirror with
    # Bone Armor; a pool that empties inside the first record; a pool on blows of the other kind only.
    for kind in (0, 1):
        for maxlife in (64 << 8, 100 << 8, 333 << 8):
            m = maxlife >> 8
            for pool in (0, 3, m // 8, m, 1000):
                frame(kind, maxlife, pool << 8, 1, [(m // 2, 0, 0, 0, 0, 2), (0, m // 2, 0, 0, 0, 0)])
                frame(kind, maxlife, pool << 8, 0, [(0, m // 2, 0, 0, 0, 0), (m // 2, 0, 0, 0, 0, 2)])
                frame(kind, maxlife, pool << 8, 1, [(m // 4, m // 4, 0, 0, 0, 2), (m // 4, 0, m // 4, m // 8, 0, 3)])
                frame(kind, maxlife, pool << 8, 0, [(m // 4, m // 4, 0, 0, 0, 2), (m // 4, 0, m // 4, m // 8, 1, 3)])
                frame(kind, maxlife, pool << 8, 1, [(0, m // 8, 0, 0, 1, 2), (m // 2, 0, 0, 0, 0, 2)])
                frame(kind, maxlife, pool << 8, 1, [(m // 16, m // 16, 0, 0, 0, 2), (m // 16, m // 16, 0, 0, 0, 2)])
    nb = len(rows)
    left = n
    while left > 0:
        kind = rng.randrange(2)
        maxlife = rng.randrange(8, 2000) << 8
        m = maxlife >> 8
        armor = rng.randrange(2)
        r = rng.random()
        pool = 0 if r < 0.1 else rng.randrange(1, m // 4 + 2) if r < 0.6 else rng.randrange(1, 4 * m)
        blows = []
        for _ in range(1 if rng.random() < 0.3 else rng.choice((2, 2, 3, 4))):
            parts = [0, 0, 0, 0]
            for k in rng.sample(range(4), rng.choice((1, 1, 2, 3))):
                parts[k] = rng.randrange(0, m // 2 + 1) if rng.random() < 0.7 else rng.randrange(0, m + 1)
            poison = 1 if rng.random() < 0.15 else 0
            hc = rng.randrange(14) if rng.random() < 0.8 else rng.choice((0x40, 0x20, 0x43, 0x12))
            blows.append((*parts, poison, hc))
        frame(kind, maxlife, pool << 8, armor, blows)
        left -= 1
    print("#", " ".join(COLS), f"boundary_records={nb}", f"random_frames={n}", f"seed={seed}")
    for row in rows:
        print("\t".join(str(v) for v in row))


def read_cases(path):
    with open(path) as f:
        for line in f:
            if line.startswith("#") or not line.strip(): continue
            yield dict(zip(COLS, (int(v) for v in line.split())))


def frames(path):
    cur = []
    for c in read_cases(path):
        cur.append(c)
        if not c["next"]:
            yield cur
            cur = []


def game(exe, path):
    import emu
    e = emu.Emu(exe) if exe else emu.Emu()
    cur = {}
    out = {}
    stats = {}
    zero_row = e.alloc(0x400)
    e.hook(0x620bb0, lambda e_: 0, 1)                                            # the unit's room: none
    e.hook(0x639df0, lambda e_: 0, 2)                                            # has state: none (frozen, 0x83, 0x85)
    e.hook(0x6256b0, lambda e_: cur["list"] if e_.arg(1) == STATE else 0, 2)     # a state's stat list: the armor's; 0x15 stun none
    e.hook(0x625d00, lambda e_: stats.get(e_.arg(1), 0) if e_.arg(0) == cur["list"] else 0, 3)
    e.hook(0x627150, lambda e_: stats.__setitem__(e_.arg(1), e_.arg(2)) if e_.arg(0) == cur["list"] else None, 4)
    e.hook(0x639e30, lambda e_: 0, 3)                                            # the state's graphic
    e.hook(0x639db0, lambda e_: out.__setitem__("removed", 1), 3)                # the state off at an empty pool
    e.hook(0x6277e0, lambda e_: 0, 2)
    e.hook(0x626cd0, lambda e_: 0, 1)
    e.hook(0x40b480, lambda e_: out.__setitem__("freed", 1), 3)                  # the node freed (not with [node+4] set)
    e.hook(0x611d30, lambda e_: zero_row, 1)                                     # the difficulty row
    e.hook(0x63ee90, lambda e_: 0, 1)                                            # no hireling class
    e.hook(0x625480, lambda e_: 0, 3)                                            # every stat 0: no resistance, reduction, absorb
    e.hook(0x625500, lambda e_: 0, 3)
    e.hook(0x63e990, lambda e_: 0, 1)
    e.hook(0x63e940, lambda e_: 0, 1)
    e.hook(0x57c060, lambda e_: 100, 0)                                          # damage percent 100: fields unscaled
    e.hook(0x46c140, lambda e_: 1 if e_.mu.reg_read(emu.UC_X86_REG_EDX) == 3 else 0, 0)   # MonStats2 GH mode bit set
    e.hook(0x625d10, lambda e_: cur["maxlife"], 1)                               # max life, 256ths
    for addr, n in ((0x5a54f0, 2), (0x5a5650, 0), (0x5734c0, 0), (0x5a43a0, 0), (0x5a7c20, 3)):
        e.hook(addr, lambda e_: 0, n)
    e.hook(0x5a7e60, lambda e_: out.__setitem__("mon_mode", e_.arg(1)), 3)
    e.hook(0x5809d0, lambda e_: out.__setitem__("pl_mode", e_.arg(1)), 5)
    e.hook(0x64c040, lambda e_: out.__setitem__("small", 1), 1)
    # The data tables the handler reads (0x744304): one skill row with the state and the two stats.
    data = e.alloc(0x1000)
    skills = e.alloc(0x23c * 4)
    e.w32(0x744304, data)
    e.w32(data + 0xba0, 4); e.w32(data + 0xb98, skills); e.w32(data + 0xc4, 0x100); e.w32(data + 0xbd4, 0x200)
    row = skills + SKILL_INDEX * 0x23c
    e.mu.mem_write(row + 0x80, STATE.to_bytes(2, "little"))
    e.mu.mem_write(row + 0x56, MAX_STAT.to_bytes(2, "little"))
    e.mu.mem_write(row + 0x54, POOL_STAT.to_bytes(2, "little"))
    game_ptr = e.alloc(0x1000)
    attacker = e.alloc(0x200)
    e.w32(attacker, 1); e.w32(attacker + 4, 0x10)
    n = 0
    for recs in frames(path):
        c0 = recs[0]
        unit = e.alloc(0x200)
        e.w32(unit, c0["kind"]); e.w32(unit + 4, 0x10); e.w32(unit + 0x10, 1)
        e.w32(unit + 0x20, c0["seedlo"]); e.w32(unit + 0x24, c0["seedhi"])
        node = e.alloc(0x40)
        e.mu.mem_write(node, b"\x0b"); e.w32(node + 4, 1); e.w32(node + 0xc, SKILL_INDEX)
        e.w32(node + 0x14, CYCLONE if c0["armor"] else BONE)
        e.w32(unit + 0x90, node)
        cur["list"] = e.alloc(0x40)
        cur["maxlife"] = c0["maxlife"]
        stats.clear(); stats[POOL_STAT] = c0["pool"]; stats[MAX_STAT] = max(c0["pool"], 1)
        gh, left, removed = 0, 0, 0
        for c in recs:
            out.clear()
            rec = e.alloc(0x100)
            e.w32(rec + 4, 1 | 4)
            e.w32(rec + 8, c["phys"]); e.w32(rec + 0x10, c["fire"]); e.w32(rec + 0x1c, c["ltng"])
            e.w32(rec + 0x24, c["cold"]); e.w32(rec + 0x28, c["poison"]); e.w32(rec + 0x60, c["hc"])
            if e.r32(unit + 0x90):                                               # the state still on (no node: the pool emptied)
                e.call(0x57c1e0, unit, rec, ecx=game_ptr, edx=attacker)
            else:
                e.w32(rec + 0x4c, c["phys"] + c["fire"] + c["ltng"] + c["cold"] + c["poison"])
            removed |= out.get("removed", 0)
            if out.get("removed"): e.w32(unit + 0x90, 0)
            total = e.s32(rec + 0x4c)
            left += total
            out.pop("removed", None)
            e.call(0x57cee0, unit, rec, ecx=game_ptr, edx=0)
            g = 1 if out.get("mon_mode") == 3 or out.get("pl_mode") == 4 else 0
            if g and out.get("small"): raise SystemExit(f"case {c['id']}: both a get-hit and too-small")
            if not g and not out.get("small"): raise SystemExit(f"case {c['id']}: neither a get-hit nor too-small: {out}")
            gh |= g
        print(c0["id"], gh, stats.get(POOL_STAT, 0), left - sum(c["poison"] for c in recs))
        n += 1
    print("#", n, "frames", file=sys.stderr)


def compare(a, b):
    def load(p):
        d = {}
        for line in open(p):
            if line.startswith("#") or not line.strip(): continue
            parts = [int(v) for v in line.split()]
            d[parts[0]] = tuple(parts[1:])
        return d
    ga, gb = load(a), load(b)
    cases = {}
    for arg in sys.argv:
        if arg.endswith(".tsv") and "case" in arg: cases = {c["id"]: c for c in read_cases(arg)}
    bad = {"gh": 0, "pool": 0, "damage": 0}
    differ, shown = 0, 0
    for i in sorted(ga):
        if i not in gb: raise SystemExit(f"case {i} missing from {b}")
        x, y = ga[i], gb[i]
        for k, name in enumerate(("gh", "pool", "damage")):
            if x[k] != y[k]: bad[name] += 1
        if x != y:
            differ += 1
            if shown < 10:
                shown += 1
                print("differ", i, "game", x, "d2d", y, cases.get(i, ""))
    print(f"compare: {len(ga)} frames, {differ} differ (get-hit {bad['gh']}, pool {bad['pool']}, damage {bad['damage']})")


def main():
    argv, exe = sys.argv[1:], None
    if "--exe" in argv:
        at = argv.index("--exe"); exe = argv[at + 1]; del argv[at:at + 2]
    if not argv: raise SystemExit(__doc__)
    if argv[0] == "gen": gen(int(argv[1]) if len(argv) > 1 else 50000, int(argv[2]) if len(argv) > 2 else 7)
    elif argv[0] == "game": game(exe, argv[1])
    elif argv[0] == "compare": compare(argv[1], argv[2])
    else: raise SystemExit(__doc__)


if __name__ == "__main__": main()
