#!/usr/bin/env python3
# SPDX-License-Identifier: GPL-3.0-or-later
"""Table-only helper of card monsterhit: what a monster melee swing is worth against a defender (no emulator).

Reads MonStats.txt and MonLvl.txt (tab separated, the MPQ's own headers; dump them with mpq-cat) and prints, per
row and difficulty, the numbers d2d's rules::monster_stats builds (V run equal to game.exe in
docs/research/monster-stats-trace.md R1): to-hit = MonLvl TH x A1TH / 100, damage = MonLvl DM x A1MinD / 100 ..
(a row with NoRatio takes the column as it is), the El1..3 side parts of the mode, and the chance to hit a defender
(rules::hit_chance, combat.hpp:28). Use it to list the rows a trace has to cover (elemental melee, A2 rows) and to
turn a defense into a hit chance and a hit per second. It proves nothing about game.exe: the rows it prints are
tagged D (derived from tables) in the trace.

    python3 mh_rows.py MonStats.txt MonLvl.txt [--diff 0|1|2] [--def 20] [--clvl 3] [--only id,id] [--el]
    python3 mh_rows.py --selftest
"""
import sys


def table(path):
    with open(path, encoding="latin-1") as f:
        lines = [l.rstrip("\r\n").split("\t") for l in f if l.strip()]
    head = lines[0]
    return [dict(zip(head, r)) for r in lines[1:]]


def num(row, key, default=0):
    v = row.get(key, "")
    try:
        return int(v)
    except ValueError:
        return default


def hit_chance(ar, defense, alvl, dlvl):
    """rules::hit_chance (combat.hpp:28)."""
    if defense < 0:
        ar, defense = ar - defense, 0
    if ar < 0:
        defense, ar = defense - ar, 0
    c = ar * 100 // (ar + defense) if ar + defense else 100
    return max(5, min(95, c * 2 * alvl // max(alvl + dlvl, 1)))


SUFFIX = ["", "(N)", "(H)"]
LEVEL_KEY = ["Level", "Level(N)", "Level(H)"]


def rows(ms, ml, diff):
    lvl = {num(r, "Level"): r for r in ml}
    for r in ms:
        if not r.get("Id") or r.get("Id") == "Expansion":
            continue
        level = num(r, LEVEL_KEY[diff])
        base = lvl.get(level)
        if base is None or num(r, "A1MaxD" + SUFFIX[diff]) + num(r, "A2MaxD" + SUFFIX[diff]) == 0:
            continue
        ratio = num(r, "NoRatio") == 0
        th, dm = num(base, "TH" + SUFFIX[diff]), num(base, "DM" + SUFFIX[diff])

        def scale(base_v, col):
            v = num(r, col + SUFFIX[diff])
            return base_v * v // 100 if ratio else v

        out = {"id": r["Id"], "level": level, "ar_a1": scale(th, "A1TH"), "ar_a2": scale(th, "A2TH"),
               "a1": (scale(dm, "A1MinD"), max(scale(dm, "A1MaxD"), scale(dm, "A1MinD"))),
               "a2": (scale(dm, "A2MinD"), max(scale(dm, "A2MaxD"), scale(dm, "A2MinD"))), "el": []}
        for i in (1, 2, 3):
            mode, typ = r.get(f"El{i}Mode", ""), r.get(f"El{i}Type", "")
            mx = num(r, f"El{i}MaxD" + SUFFIX[diff])
            if typ and mx > 0:
                out["el"].append((mode, typ, num(r, f"El{i}Pct" + SUFFIX[diff]) or 100,
                                  scale(dm, f"El{i}MinD"), max(scale(dm, f"El{i}MaxD"), scale(dm, f"El{i}MinD")),
                                  num(r, f"El{i}Dur" + SUFFIX[diff])))
        yield out


def show(ms, ml, diff, defense, clvl, only, el_only):
    print("id lvl | AR a1/a2 | A1 min-max | A2 min-max | chance a1 vs DEF | mean hit | El (mode type pct min-max dur)")
    for o in rows(ms, ml, diff):
        if only and o["id"] not in only:
            continue
        if el_only and not o["el"]:
            continue
        chance = hit_chance(o["ar_a1"], defense, o["level"], clvl)
        mean = (o["a1"][0] + o["a1"][1]) / 2
        print(o["id"], o["level"], "|", o["ar_a1"], o["ar_a2"], "|", "%d-%d" % o["a1"], "|", "%d-%d" % o["a2"], "|",
              chance, "% |", round(mean, 1), "|", o["el"])


def selftest():
    ms = [{"Id": "fallen1", "Level": "1", "Level(N)": "36", "Level(H)": "69", "A1MinD": "50", "A1MaxD": "100", "A1TH": "100",
           "A1MinD(N)": "50", "A1MaxD(N)": "100", "A1TH(N)": "100", "A1MinD(H)": "50", "A1MaxD(H)": "100", "A1TH(H)": "100",
           "A2MaxD": "0", "NoRatio": "0", "El1Mode": "A1", "El1Type": "fire", "El1Pct": "0", "El1MinD": "10", "El1MaxD": "20", "El1Dur": "0"}]
    ml = [{"Level": "1", "TH": "8", "DM": "2", "TH(N)": "0", "DM(N)": "0", "TH(H)": "0", "DM(H)": "0"},
          {"Level": "36", "TH": "0", "DM": "0", "TH(N)": "100", "DM(N)": "40", "TH(H)": "0", "DM(H)": "0"}]
    n = list(rows(ms, ml, 0))[0]
    assert n["ar_a1"] == 8 and n["a1"] == (1, 2) and n["el"] == [("A1", "fire", 100, 0, 0, 0)], n
    h = list(rows(ms, ml, 1))[0]
    assert h["level"] == 36 and h["ar_a1"] == 100 and h["a1"] == (20, 40), h
    # 8 AR vs DEF 20 at monster level 1 vs clvl 3: 8*100//28 = 28; 28*2*1//4 = 14 (archive doc: 14 percent)
    assert hit_chance(8, 20, 1, 3) == 14
    assert hit_chance(1000, 0, 50, 1) == 95 and hit_chance(1, 1000, 1, 90) == 5
    print("selftest ok")


def main():
    a = sys.argv[1:]
    if a == ["--selftest"]:
        return selftest()
    if len(a) < 2:
        sys.exit(__doc__)
    opts = {"--diff": "0", "--def": "20", "--clvl": "3", "--only": ""}
    rest, i = [], 2
    while i < len(a):
        if a[i] in opts:
            opts[a[i]] = a[i + 1]
            i += 2
        else:
            rest.append(a[i])
            i += 1
    show(table(a[0]), table(a[1]), int(opts["--diff"]), int(opts["--def"]), int(opts["--clvl"]),
         set(filter(None, opts["--only"].split(","))), "--el" in rest)


main()
