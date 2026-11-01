# SPDX-License-Identifier: GPL-3.0-or-later
"""SKGROUND2 D4: where a standing missile hits a monster, game.exe vs a real d2d build.
    python sk_gnd2_hit_cmp.py game.txt d2d.txt
game.txt = `sk_gnd2_hit.py SPAN ROW FRAMES` (`H size class a b hit first count`),
d2d.txt = `test_sk_ground --hit ROW FRAMES` (`H size class dx dy hits`). Compared per
monster class that both ran (same MonStats row), offsets -SPAN..SPAN subtiles."""
import sys


def load(path, game):
    out = {}
    for ln in open(path):
        p = ln.split()
        if not p or p[0] != "H" or len(p) < 6:
            continue
        if game and p[5] in ("-", "nomissile"):
            continue
        cls, a, b = int(p[2]), int(p[3]), int(p[4])
        n = int(p[7]) if game and len(p) > 7 else (int(p[5]) if not game else int(p[5]))
        out.setdefault(cls, {})[(a, b)] = (int(p[1]), n)
    return out


g, d = load(sys.argv[1], True), load(sys.argv[2], False)
mism = 0
for cls in sorted(set(g) & set(d)):
    size = next(iter(g[cls].values()))[0]
    gs = {k for k, (_, n) in g[cls].items() if n}
    ds = {k for k, (_, n) in d[cls].items() if n}
    cells = set(g[cls]) & set(d[cls])
    gs &= cells
    ds &= cells
    diff = gs ^ ds
    mism += len(diff)
    gt = sum(g[cls][k][1] for k in gs & ds)
    dt = sum(d[cls][k][1] for k in gs & ds)
    print(f"class {cls} (unit size {size}): game.exe hits {len(gs)} offsets, d2d {len(ds)}; "
          f"only game.exe {len(gs - ds)}, only d2d {len(ds - gs)}; hits on the {len(gs & ds)} shared offsets: game.exe {gt}, d2d {dt}")
print(f"MISMATCH offsets {mism} (cells {len(cells)} per class)")
