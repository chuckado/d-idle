# SPDX-License-Identifier: GPL-3.0-or-later
"""Verifier (card verifymods): sort the field mismatches of verify_mods.py (game.exe) against
test_verify_mods --verify-mods (d2d) into the candidate's claims and the known leftovers.

    python verify_mods_compare.py game.txt d2d.txt

Only the case lines (`... | ...`) count; d2d's log lines are skipped. A field mismatch is put in:
  life-round   life off by 1 or 2 (FUN_005a0d20's 8.8 fixed point; d2d keeps whole points)
  B            a berserker champion (39): level, experience
  D3           a ghostly champion (36): velocity
  D2           a minion: level, experience
  D1           life (off by more than 2), defence or experience of a boss
  stoneskin    defence of a unique with mod 28
  other        anything else (printed, up to 20)
"""
import collections
import sys

FIELDS = ("level", "life", "def", "exp", "vel")


def lines(path):
    out = {}
    for line in open(path, encoding="latin1"):
        if " | " not in line: continue
        head, _, tail = line.rstrip("\n").partition(" | ")
        out[head] = tail.split()
    return out


def main():
    game, d2d = lines(sys.argv[1]), lines(sys.argv[2])
    counts, cases_bad, shown = collections.Counter(), 0, 0
    for head, g in game.items():
        d = d2d.get(head)
        if d is None:
            counts["missing"] += 1
            continue
        if g == d: continue
        cases_bad += 1
        if g == ["-"] or d == ["-"]:
            counts["placement"] += 1
            continue
        _, _, _, kind, mods, _ = head.split()
        mods = mods.split(".")
        for name, a, b in zip(FIELDS, g, d):
            if a == b: continue
            a, b = int(a), int(b)
            if name == "life" and abs(a - b) <= 2: slot = "life-round"
            elif kind == "c" and "39" in mods and name in ("level", "exp", "life"): slot = "B"
            elif kind == "c" and "36" in mods and name == "vel": slot = "D3"
            elif kind == "m" and name in ("level", "exp", "life", "def"): slot = "D2"
            elif name == "def" and "28" in mods: slot = "stoneskin"
            elif name in ("life", "def", "exp"): slot = "D1"
            else:
                slot = "other"
                if shown < 20:
                    shown += 1
                    print("other:", head, "|", " ".join(g), "vs", " ".join(d))
            counts[slot] += 1
    print(f"cases {len(game)}, cases differing {cases_bad}")
    for slot in ("D1", "D2", "D3", "B", "life-round", "stoneskin", "placement", "missing", "other"):
        print(f"  {slot:10s} {counts[slot]}")


if __name__ == "__main__":
    main()
