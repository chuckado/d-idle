# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe 1.14d's monster stat rows (FUN_006538a0, called by FUN_00573cb0) as the oracle (agent monsterstats).

    tools/emu/ms_run.sh ms_stats.py game  > game.tsv   # every row x difficulty x level 1..last MonLvl row
    tools/emu/ms_run.sh ms_stats.py flags              # which output slots each flag bit of the call fills
    ms_d2d_test --stats > d2d.tsv                       # tools/emu/ms_d2d_test.cpp (d2d's real monster_stats / monster_ranges)
    python3 ms_stats.py cmp game.tsv d2d.tsv            # field-by-field comparison

`game` lines: `row diff level minhp maxhp ac th a1min a1max a2min a2max s1min s1max exp th_a2 th_s1`
(FUN_006538a0(row, 1, diff, level, 0x3f, out): eleven int32 at out, the order of upstream's
monstats.py). Nothing is hooked: the function reads MonStats / MonLvl as the loader built them.
"""
import sys


def game(e, flags=0x3f):
    import emu
    from monstats import table
    monstats = [r for r in table("monstats") if r.get("id") and r["id"] != "Expansion"]
    monlvl = table("monlvl")
    out = e.alloc(0x40)
    for row in range(len(monstats)):
        for diff in range(3):
            for level in range(1, len(monlvl)):
                # flag bits (V run, `flags`): 1 life, 2 defence, 4 experience, 8 A1 (to-hit + damage), 0x10 A2, 0x20 S1
                got = []
                for fl in (0x0f, 0x10, 0x20):
                    e.mu.mem_write(out, bytes(0x40))
                    ok = e.call(0x6538a0, row, 1, diff, level, fl, out)
                    got.append([e.s32(out + 4 * i) for i in range(11)] if ok else None)
                if None in got:
                    print(monstats[row]["id"].lower(), diff, level, "none"); continue
                a, b, c = got
                print(monstats[row]["id"].lower(), diff, level, a[0], a[1], a[2], a[3], a[5], a[6], b[7], b[8], c[9], c[10], a[4], b[3], c[3])


def flags(e):
    from monstats import table
    monstats = [r for r in table("monstats") if r.get("id") and r["id"] != "Expansion"]
    out = e.alloc(0x40)
    row = next(i for i, r in enumerate(monstats) if r["id"] == "fallen1")
    for flag in (0, 1, 2, 4, 8, 0x10, 0x20, 0x3f):
        e.mu.mem_write(out, bytes(0x40))
        ok = e.call(0x6538a0, row, 1, 0, 5, flag, out)
        print(f"flag {flag:#x} ret {ok}:", *[e.s32(out + 4 * i) for i in range(11)])


def players(e):
    """FUN_005738f0 (life %) and FUN_00573910 (experience %) for 0..16 players, fastcall ecx = the count:
    lines `n life exp` like `ms_d2d_test --players`."""
    for n in range(17):
        print(n, e.call(0x5738f0, ecx=n) & 0xffffffff, e.call(0x573910, ecx=n) & 0xffffffff)


def probe(e):
    """FUN_006538a0 for class 19 (a Fallen in the Blood Moor trace), flags 7 as FUN_00573cb0 passes, over the
    expansion flag (arg 2), difficulty (arg 3) and level 36 / 5: which combination gives the unit's 314?"""
    out = e.alloc(0x40)
    for exp in (0, 1):
        for diff in (0, 1, 2):
            for level in (5, 36):
                e.mu.mem_write(out, bytes(0x40))
                ok = e.call(0x6538a0, 19, exp, diff, level, 7, out)
                print(f"row 19 expansion {exp} diff {diff} level {level} ret {ok}:", *[e.s32(out + 4 * i) for i in range(5)])


def cmp(game_file, d2d_file, meta_file):
    # rows are matched by (id, n-th row with that id): both lists keep the MonStats order
    ids, seen = {}, {}
    for l in open(meta_file):
        t = l.split()
        if len(t) >= 5 and t[0].isdigit():
            seen[t[1].lower()] = seen.get(t[1].lower(), 0) + 1
            ids[int(t[0])] = (t[1].lower(), seen[t[1].lower()])
    names = "minhp maxhp ac th a1min a1max a2min a2max s1min s1max exp th2 ths1".split()
    g, seen = {}, {}
    for line in open(game_file):
        f = line.split()
        if len(f) < 4 or not f[1].isdigit(): continue
        if f[1] == "0" and f[2] == "1": seen[f[0]] = seen.get(f[0], 0) + 1
        if f[3] == "none": continue
        g[(f[0], seen[f[0]], f[1], f[2])] = [int(x) for x in f[3:]]
    cases = diff = unmatched = zero_life = 0
    per = {n: 0 for n in names}
    per["hp_roll_outside"] = 0
    rows = {}
    shown = 0
    for line in open(d2d_file):
        f = line.split()
        if len(f) != 18 or not f[0].isdigit() or int(f[0]) not in ids: continue
        idn = ids[int(f[0])]
        key = (idn[0], idn[1], f[1], f[2])
        if key not in g: unmatched += 1; continue
        w = g[key]
        d = [int(x) for x in f[3:]]
        # d2d: minhp maxhp ac th a1min a1max a2min a2max s1min s1max exp lo32 hi32 th2 ths1
        dd = d[:11] + d[13:15]
        cases += 1
        bad = [names[i] for i in range(13) if dd[i] != w[i]]
        # d2d never rolls less than 1 hit point (max(.., 1)); a game row with min = max = 0 life is not a fighting monster
        roll = (d[11] < w[0] or d[12] > w[1]) and not (w[0] == 0 and w[1] == 0 and d[11] == 1 and d[12] == 1)
        if w[0] == 0 and w[1] == 0: zero_life += 1
        for n in bad: per[n] += 1
        if roll: per["hp_roll_outside"] += 1
        if bad or roll:
            diff += 1
            rows.setdefault(idn, []).append((f[1], int(f[2]), bad, roll))
            if shown < 6:
                shown += 1; print("diff", key, "game", w, "d2d", dd, d[11:13], bad)
    print(f"{cases} cases, {diff} differ, {unmatched} d2d lines with no game row, {zero_life} cases with zero game life; per field: {per}; game keys {len(g)}")
    for idn, v in sorted(rows.items()):
        print("  row", idn, len(v), "cases; fields", sorted({b for _, _, bad, _ in v for b in bad}), "roll" if any(r for *_, r in v) else "", "diff/levels", sorted({(x[0], x[1]) for x in v})[:3], "..")


if __name__ == "__main__":
    mode = sys.argv[1]
    if mode == "cmp":
        cmp(sys.argv[2], sys.argv[3], sys.argv[4])
    else:
        import drlg
        e = drlg.boot()
        {"game": game, "flags": flags, "probe": probe, "players": players}[mode](e)
