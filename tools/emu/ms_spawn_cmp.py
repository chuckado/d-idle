# SPDX-License-Identifier: GPL-3.0-or-later
"""Compare the real unit stats game.exe gave (ms_spawn.py) with d2d's rows (ms_d2d_test) (agent monsterstats).

    python3 ms_spawn_cmp.py DIFF spawn.tsv d2d.tsv meta.tsv res.tsv bin.tsv

Per plain monster (flags & 0x1c == 0) of the spawn trace: its class id -> d2d row (the loader's
mon_bin), then: level stat 12 / the level FUN_006538a0 read = d2d's type level for the difficulty;
life in [d2d min, d2d max] at that level; experience stat 13 = d2d exp; defence stat 0x1f = d2d
armor_class; the six base resistances (stats 0x24 0x25 0x27 0x29 0x2b 0x2d) = d2d's MonStats columns.
Also reports the life range actually seen per row (both ends reached or not).
"""
import sys


def rd(path, n):
    for line in open(path):
        f = line.split()
        if len(f) >= n and f[0].isdigit(): yield f


def main():
    diff = int(sys.argv[1])
    spawn, d2d, meta, res, binf = sys.argv[2:7]
    row_of = {int(f[0]): int(f[1]) for f in rd(binf, 2)}
    level = {int(f[0]): int(f[5]) if diff == 0 else int(f[2 + diff - 1]) for f in rd(meta, 6)}
    ids = {int(f[0]): f[1] for f in rd(meta, 6)}
    resists = {(int(f[0]), int(f[1])): [int(x) for x in f[2:8]] for f in rd(res, 8)}
    need = {}
    for f in rd(d2d, 18):
        if int(f[1]) == diff: need[(int(f[0]), int(f[2]))] = f
    cnt = dict(units=0, plain=0, level=0, lvlstat=0, hp_out=0, exp=0, ac=0, res=0, nolookup=0)
    seen = {}
    bad = {}
    shown_flag = [0]
    sums = [0.0] * 8    # units, game exp, d2d-level exp, game life, d2d-level mid life; units at another level, their game exp, d2d exp
    for line in open(spawn):
        f = line.split()
        if not f or f[0] != "U": continue
        cnt["units"] += 1
        lid, seed, n, cls, flags, statlvl = f[1], f[2], f[3], int(f[4]), int(f[5], 16), f[6]
        if flags & 0x1c: continue
        cnt["plain"] += 1
        before = sum(cnt[k] for k in ("level", "hp_out", "exp", "ac", "res"))
        if shown_flag[0]: pass
        st = {}
        for tok in f[8:]:
            k, _, v = tok.partition(":")
            st[int(k, 16)] = int(v)
        row = row_of.get(cls)
        if row is None or row not in level or statlvl == "None":
            cnt["nolookup"] += 1; continue
        lv = st.get(0xc, 0)
        if int(statlvl) != lv:
            # rows read at the base level, stat 12 raised afterwards (leveladd): a boss-pack member whose
            # MonsterData flag byte is still 0 (seen: champion packs' members). Not a plain monster.
            cnt["boosted_unflagged"] = cnt.get("boosted_unflagged", 0) + 1
            continue
        g = need.get((row, lv))
        if g is None: cnt["nolookup"] += 1; continue
        mn, mx, ac, exp = int(g[3]), int(g[4]), int(g[5]), int(g[13])
        own = need.get((row, level[row]))
        if own is not None:
            # what d2d would have given at ITS level (MonStats Level column) against the game's level
            sums[0] += 1; sums[1] += st.get(0xd, 0); sums[2] += int(own[13]); sums[3] += st.get(7, 0) / 256
            sums[4] += (int(own[3]) + int(own[4])) / 2
            if lv != level[row]: sums[5] += 1; sums[6] += st.get(0xd, 0); sums[7] += int(own[13])
        hp = st.get(7, 0) / 256
        key = ids[row]
        r = seen.setdefault(key, [1 << 30, 0, set()]); r[0] = min(r[0], hp); r[1] = max(r[1], hp); r[2].add(lv)
        if lv != level[row] or int(statlvl) != lv:
            cnt["level"] += 1; bad.setdefault(("level", key, lv, level[row]), 0); bad[("level", key, lv, level[row])] += 1
        if not (mn <= hp <= max(mx, mn)):
            cnt["hp_out"] += 1; bad.setdefault(("hp", key, hp, mn, mx), 0); bad[("hp", key, hp, mn, mx)] += 1
        if st.get(0xd, 0) != exp:
            cnt["exp"] += 1; bad.setdefault(("exp", key, st.get(0xd, 0), exp), 0); bad[("exp", key, st.get(0xd, 0), exp)] += 1
        if st.get(0x1f, 0) != ac:
            cnt["ac"] += 1; bad.setdefault(("ac", key, st.get(0x1f, 0), ac), 0); bad[("ac", key, st.get(0x1f, 0), ac)] += 1
        want = resists[(row, diff)]
        got = [st.get(s, 0) for s in (0x24, 0x25, 0x27, 0x29, 0x2b, 0x2d)]
        if got != want:
            cnt["res"] += 1; bad.setdefault(("res", key, tuple(got), tuple(want)), 0); bad[("res", key, tuple(got), tuple(want))] += 1
        if sum(cnt[k] for k in ("level", "hp_out", "exp", "ac", "res")) != before and shown_flag[0] < 6:
            shown_flag[0] += 1; print("example:", line.strip())
    print(cnt)
    for k, v in sorted(bad.items(), key=lambda kv: -kv[1])[:25]: print("  ", k, v)
    if sums[0]:
        print(f"d2d-level view: {int(sums[0])} units; mean exp game {sums[1] / sums[0]:.1f} vs d2d level {sums[2] / sums[0]:.1f}; mean life game {sums[3] / sums[0]:.1f} vs d2d mid {sums[4] / sums[0]:.1f}")
    if sums[5]:
        print(f"units at another level than d2d's: {int(sums[5])}; their mean exp game {sums[6] / sums[5]:.1f} vs d2d {sums[7] / sums[5]:.1f}")
    print("rows seen:", len(seen))
    for k, (lo, hi, lv) in sorted(seen.items()):
        mn = None
        print(f"  {k}: life {lo:g}..{hi:g} levels {sorted(lv)}")


main()
