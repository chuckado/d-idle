# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND2: a real d2d build's Blizzard shards against game.exe's offsets (standalone copy of the
`d2d` mode that the PR candidate's tools/emu/sk_ground_blizz.py carries).

    python sk_gnd2_blizz_d2d.py cases game.tsv d2d.txt

cases / game.tsv come from `sk_ground_blizz.py gen` / `game` (SKGROUND); d2d.txt is
`test_sk_ground --blizz SEED N` (lines `S<case> <frame> <dx> <dy>`).
"""
import sys

MASK = 0xFFFFFFFF


def cases(path):
    with open(path) as f:
        for ln in f:
            p = ln.split()
            if len(p) == 6:
                yield tuple(int(v) for v in p)


def d2d(cpath, gpath, dpath):
    """d2d.txt = `test_sk_ground --blizz SEED N` (lines `S<case> <frame> <dx> <dy>`, radius 7 = Blizzard).
    game.exe's offset set at radius 7 comes from the `game` run of the same cases file (made shards only)."""
    g = {}
    with open(gpath) as f:
        for ln in f:
            p = ln.split()
            if len(p) == 4 and p[0].isdigit() and p[1] == "1":
                g[int(p[0])] = (int(p[2]), int(p[3]))
    allowed = set()
    for cid, cx, cy, age, freq, radius in cases(cpath):
        if radius == 7 and cid in g:
            allowed.add(g[cid][0] - cx)
            allowed.add(g[cid][1] - cy)
    n = bad = 0
    made = set()
    per_cast = {}
    with open(dpath) as f:
        for ln in f:
            p = ln.split()
            if len(p) == 4 and p[0].startswith("S"):
                n += 1
                dx, dy = int(p[2]), int(p[3])
                made.update((dx, dy))
                per_cast[p[0]] = per_cast.get(p[0], 0) + 1
                if dx not in allowed or dy not in allowed:
                    bad += 1
    print("game.exe offsets at radius 7 (%d): %d..%d" % (len(allowed), min(allowed), max(allowed)))
    print("d2d offsets (%d): %d..%d; never made by d2d: %s" % (len(made), min(made), max(made), sorted(allowed - made)))
    print("d2d shards %d in %d casts; shards with an offset game.exe never makes: %d (%.2f%%)"
          % (n, len(per_cast), bad, 100.0 * bad / max(n, 1)))
    print("MISMATCH %d" % bad)


if __name__ == "__main__":
    d2d(sys.argv[1], sys.argv[2], sys.argv[3])
