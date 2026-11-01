# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND: where Blizzard's shards fall, by game.exe's own code.

    python sk_ground_blizz.py gen N SEED > cases     # grid + N random cases
    python sk_ground_blizz.py game cases > game.tsv  # FUN_005a9820 natively under the emulator
    python sk_ground_blizz.py cmp cases game.tsv     # vs the reading (port) and vs d2d's rule

Runs next to upstream's tools/emu (emu.py), like sk_resist.py.

FUN_005a9820 is the shard spawner of missile srvdofunc 10 (FUN_005aea60,
blizzardcenter; also do 25): EAX = calc2 (frequency), ESI = the centre
missile, stack (game, calc1 radius, SubMissile1 row, collision mask 5).
Native: the seed set FUN_00650e40, the draws FUN_0045c3e0, the memset.
Hooked (not the claim): the age FUN_0064a3b0 (Range - Range left), the owner
FUN_00552fd0, the room FUN_00620bb0, the centre's subtile FUN_0045adf0 /
FUN_0045ae20, skill / level FUN_0064a280 / FUN_0064a210, the spot's
collision FUN_0064cb30 (always free here) and the missile maker FUN_0059fa30
(records the params block's start subtile, makes nothing).

A case is `id cx cy age freq radius`. `game` prints `id made x y` (made 0:
no shard this frame). `cmp` checks every line against
  reading: made = age % freq == 0; seed {cx + age, 666}; r = radius - 1;
           x = cx + rand(2r) - r, y = cy + rand(2r) - r (x drawn first);
           rand(n) = 0 for n < 1, else the seed step's low % n
  d2d:     Fight::missile_tick (components/game/fight/missiles.cpp, do 10 / 25):
           offset = rng(2 radius - 1) - (radius - 1) on each axis, the fight's Rng.
d2d's draw is on another generator, so the d2d check is on the RANGE of the
offsets only (values game.exe can never produce), not draw for draw.
"""
import os
import sys

sys.path.insert(0, os.getcwd())

MASK = 0xFFFFFFFF


def xorshift(state):
    x = state[0]
    x ^= (x << 13) & MASK
    x ^= x >> 17
    x ^= (x << 5) & MASK
    state[0] = x & MASK
    return state[0]


def gen(n, seed):
    st = [seed & MASK or 1]
    i = 0
    for radius in (1, 2, 3, 7, 8, 12, 20):
        for age in range(0, 100):
            for cx, cy in ((5000, 5000), (5123, 4877), (25000, 5600)):
                print(i, cx, cy, age, 4, radius)
                i += 1
    for _ in range(n):
        cx, cy = 1000 + xorshift(st) % 60000, 1000 + xorshift(st) % 60000
        age, freq = xorshift(st) % 100, 1 + xorshift(st) % 6
        radius = (7, 7, 7, 1 + xorshift(st) % 25)[xorshift(st) % 4]
        print(i, cx, cy, age, freq, radius)
        i += 1


def cases(path):
    with open(path) as f:
        for ln in f:
            p = ln.split()
            if len(p) == 6:
                yield tuple(int(v) for v in p)


def game(path):
    import emu
    from unicorn.x86_const import UC_X86_REG_EDX
    e = emu.Emu()
    cur = {}
    made = []

    def ret(v):
        return lambda _e: v
    e.hook(0x64a3b0, lambda _e: cur["age"], 1)
    e.hook(0x552fd0, ret(0x1000), 0)
    e.hook(0x620bb0, ret(0x2000), 1)
    e.hook(0x45adf0, lambda _e: cur["cx"], 0)
    e.hook(0x45ae20, lambda _e: cur["cy"], 0)
    e.hook(0x64a280, ret(59), 1)
    e.hook(0x64a210, ret(20), 1)
    e.hook(0x64cb30, ret(0), 4)

    def on_make(e):
        p = e.mu.reg_read(UC_X86_REG_EDX)
        made.append((e.s32(p + 0x14), e.s32(p + 0x18), e.r32(p + 0x10)))
        return 0
    e.hook(0x59fa30, on_make, 0)
    unit = e.alloc(0x200)
    e.w32(unit, 3)                                   # a missile unit
    for cid, cx, cy, age, freq, radius in cases(path):
        cur.update(age=age, cx=cx, cy=cy)
        made.clear()
        e.call(0x5a9820, 1, radius, 159, 5, regs={"eax": freq, "esi": unit})
        if made:
            print(cid, 1, made[0][0], made[0][1])
        else:
            print(cid, 0, "-", "-")


def rand(seed, n):
    lo, hi = seed
    v = lo * 0x6AC690C5 + hi
    seed[0], seed[1] = v & MASK, (v >> 32) & MASK
    if n < 1:
        return 0
    return seed[0] % n


def reading(cx, cy, age, freq, radius):
    if freq == 0 or age % freq:
        return (0, None, None)
    seed = [(cx + age) & MASK, 666]
    r = radius - 1
    dx = rand(seed, 2 * r) - r
    dy = rand(seed, 2 * r) - r
    return (1, cx + dx, cy + dy)


def cmp(cpath, gpath):
    g = {}
    with open(gpath) as f:
        for ln in f:
            p = ln.split()
            if len(p) == 4 and p[0].isdigit():
                g[int(p[0])] = (int(p[1]), None if p[2] == "-" else int(p[2]), None if p[3] == "-" else int(p[3]))
    n = bad = made = outside_d2d = 0
    d2d_only = {}
    seen_off = {}
    for cid, cx, cy, age, freq, radius in cases(cpath):
        n += 1
        got = g.get(cid)
        want = reading(cx, cy, age, freq, radius)
        if got != want:
            bad += 1
            if bad <= 5:
                print("reading differs", cid, got, want)
        if got and got[0]:
            made += 1
            dx, dy = got[1] - cx, got[2] - cy
            lo, hi = -(radius - 1), radius - 1          # d2d: rng(2R-1) - (R-1)
            if not (lo <= dx <= hi and lo <= dy <= hi):
                outside_d2d += 1
            s = seen_off.setdefault(radius, set())
            s.add(dx)
            s.add(dy)
    for radius, s in sorted(seen_off.items()):
        d2d_vals = set(range(-(radius - 1), radius))
        d2d_only[radius] = sorted(d2d_vals - s)
        print("radius %d: game.exe offsets %d..%d (%d values); d2d %d..%d (%d values); d2d-only %s"
              % (radius, min(s), max(s), len(s), -(radius - 1), radius - 1, len(d2d_vals), d2d_only[radius]))
    print("cases %d, shards made %d, reading differs %d, game offset outside d2d's range %d" % (n, made, bad, outside_d2d))
    for radius in (7,):
        if radius in seen_off:
            k = 2 * radius - 1
            print("radius %d: share of d2d shards with an offset game.exe never makes: 1 - (%d/%d)^2 = %.4f"
                  % (radius, k - 1, k, 1 - ((k - 1) / k) ** 2))


if __name__ == "__main__":
    what = sys.argv[1] if len(sys.argv) > 1 else ""
    if what == "gen":
        gen(int(sys.argv[2]), int(sys.argv[3]))
    elif what == "game":
        game(sys.argv[2])
    elif what == "cmp":
        cmp(sys.argv[2], sys.argv[3])
    else:
        sys.exit(__doc__)
