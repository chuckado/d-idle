#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""The gear's elemental damage of a player's melee blow: game.exe 1.14d FUN_0057a8e0 / FUN_0057a880 vs d2d make_fighter + player_blow (card decomp35).

    python d35_elem.py gen N SEED > cases.tsv    # id seed0 seed1 element min max mastery
    python d35_elem.py game cases.tsv > game.tsv # id value256   (the record's field after the roll, 256ths)
    d35_elem_harness < cases.tsv > d2d.tsv       # id whole points d2d rolls
    python d35_elem.py cmp cases.tsv game.tsv d2d.tsv
element 0 fire (stats 0x30 / 0x31, mastery 0x149), 1 lightning (0x32 / 0x33, 0x14a), 2 cold (0x36 / 0x37, 0x14b), 4 magic (0x34 / 0x35, 0x165).
min / max in whole points, mastery in percent. The roll is FUN_0045c3e0 on the unit's seed (+0x20 / +0x24), both sides start from the same two words.
"""
import random, sys
M32 = 0xFFFFFFFF
COLS = "id s0 s1 el mn mx mast".split()
PAIR = {0: (0x30, 0x31, 0x149), 1: (0x32, 0x33, 0x14a), 2: (0x36, 0x37, 0x14b), 4: (0x34, 0x35, 0x165)}


def gen(n, seed):
    rnd = random.Random(seed); rows = []
    def add(el, mn, mx, mast): rows.append((len(rows), rnd.getrandbits(32), rnd.getrandbits(32), el, mn, mx, mast))
    for el in (0, 1, 2, 4):
        for mn, mx in ((1, 1), (1, 2), (0, 1), (2, 5), (5, 5), (10, 20), (1, 40), (30, 30), (0, 0), (3, 9), (15, 25), (50, 100)):
            for mast in (0, 1, 5, 10, 25, 50, 100, 200):
                for _ in range(3): add(el, mn, mx, mast)
    nb = len(rows)
    for _ in range(n):
        mn = rnd.choice((0, 1, 2, 3, 5, 8, 10, 15, rnd.randrange(0, 60))); mx = mn + rnd.choice((0, 0, 1, 2, 5, 10, rnd.randrange(0, 40)))
        add(rnd.choice((0, 1, 2, 4)), mn, mx, rnd.choice((0, 0, 0, 5, 10, 20, 30, 50, 80, 100, rnd.randrange(0, 200))))
    print("#", " ".join(COLS), f"boundary={nb}", f"random={n}", f"seed={seed}")
    for r in rows: print("\t".join(map(str, r)))


A = 0x6AC690C5


def step(s0, s1):
    """d2d Rng::next (FUN_0045c3e0's draw): the state after the to-hit roll d2d draws first."""
    x = s0 * A + s1
    return x & M32, x >> 32


def read_cases(path):
    for line in open(path):
        if line.startswith("#") or not line.strip(): continue
        yield dict(zip(COLS, (int(v) for v in line.split())))


def game(exe, path):
    import emu
    e = emu.Emu(exe) if exe else emu.Emu()
    st = {}
    att = e.alloc(0x100)
    e.hook(0x625480, lambda e_: st.get((e_.arg(0), e_.arg(1)), 0), 3)
    for c in read_cases(path):
        st.clear()
        mn_id, mx_id, ma_id = PAIR[c["el"]]
        st[(att, mn_id)] = c["mn"]; st[(att, mx_id)] = c["mx"]; st[(att, ma_id)] = c["mast"]
        g0, g1 = step(c["s0"], c["s1"])
        e.w32(att + 0x20, g0); e.w32(att + 0x24, g1)
        v = e.call(0x57a8e0, mn_id, ma_id, 0, 0, 0, regs={"eax": mx_id, "edi": att})
        print(c["id"], v if v < 1 << 31 else v - (1 << 32))


def cmp(cpath, gpath, dpath):
    cases = list(read_cases(cpath))
    g = [list(map(int, l.split())) for l in open(gpath) if not l.startswith("#")]
    d = [list(map(int, l.split())) for l in open(dpath) if not l.startswith("#")]
    whole = fine = n = 0; sg = sd = 0
    for c, gl, dl in zip(cases, g, d):
        if dl[1] == 0: continue                                    # a to-hit miss on d2d's first draw
        n += 1
        whole += (gl[1] >> 8) != dl[2]; fine += gl[1] != dl[2] << 8
        sg += gl[1]; sd += dl[2] << 8
    print(f"cmp: {n} cases; whole points differ in {whole}; 256ths differ in {fine}; mean game {sg / n / 256:.4f} d2d {sd / n / 256:.4f}")
    return 0 if fine == 0 else 1


if __name__ == "__main__":
    argv, exe = sys.argv[1:], None
    if "--exe" in argv:
        k = argv.index("--exe"); exe = argv[k + 1]; del argv[k:k + 2]
    if argv[0] == "gen": gen(int(argv[1]), int(argv[2]))
    elif argv[0] == "game": game(exe, argv[1])
    elif argv[0] == "cmp": sys.exit(cmp(argv[1], argv[2], argv[3]))
