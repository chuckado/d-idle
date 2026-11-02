#!/usr/bin/env python3
"""mercstats: our bot's hire and revive formulas (internal/merc/formulas.go, internal/resource/merc.go, written out here the
same way) against game.exe's own offers and prices.

    python3 -I tools/emu/mst_botcost.py hireling.txt gameoffers.txt [gameoffers2.txt ..] revive_game.txt

hireling.txt: the table of `mpq-table.sh hireling`; gameoffers: `mst_oracle.py hire A-B` output; revive_game.txt:
`mst_oracle.py revive` output. Prints the case counts and mismatches (bot formula vs game.exe) for the hire cost and the
revive price. The Go code is read, not run (the Go unit tests of the package pass separately).
"""
import re
import sys


def table(path):
    lines = [l.rstrip("\n").split("\t") for l in open(path) if not l.startswith("labrun")]
    head = lines[0]
    return head, lines[1:]


def hire_cost(base_gold, offer_level, row_level):            # internal/merc/formulas.go HireCost
    d = max(0, offer_level - row_level)
    return max(base_gold, base_gold * (15 * d + 100) // 100)


def revive_cost_merc(level):                                  # internal/merc/formulas.go ReviveCost
    level = max(level, 1)
    return min(level * level // 2 * 15, 50000)


def revive_cost_resource(level):                              # internal/resource/merc.go MercReviveCost
    return 0 if level <= 0 else min(level * level // 2 * 15, 50000)


def main():
    head, rows = table(sys.argv[1])
    ix = {n: head.index(n) for n in ("Version", "Id", "Level", "Gold")}
    first = {}
    for r in rows:
        if len(r) > ix["Gold"] and r[ix["Version"]] == "100":
            i = int(r[ix["Id"]])
            if i not in first or int(r[ix["Level"]]) < first[i][1]:
                first[i] = (int(r[ix["Gold"]]), int(r[ix["Level"]]))
    cases = bad = 0
    shown = 0
    for path in sys.argv[2:-1]:
        for line in open(path):
            m = re.match(r"^[0-9a-f]{8} \d+ \d+ \d+: (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) ", line)
            if not m:
                continue
            ident, level, cost = int(m.group(1)), int(m.group(2)), int(m.group(6))
            gold, row_level = first[ident]
            cases += 1
            if hire_cost(gold, level, row_level) != cost:
                bad += 1
                if shown < 5:
                    shown += 1
                    print("hire mismatch:", line.strip(), "bot", hire_cost(gold, level, row_level))
    print(f"hire cost: {cases} offers, bot formula differs from game.exe in {bad}")
    n = bad_m = bad_r = 0
    for line in open(sys.argv[-1]):
        m = re.match(r"^(\d+) (\d+)$", line)
        if not m:
            continue
        level, price = int(m.group(1)), int(m.group(2))
        n += 1
        if level >= 1 and revive_cost_merc(level) != price:
            bad_m += 1
        if level >= 1 and revive_cost_resource(level) != price:
            bad_r += 1
    print(f"revive price: {n} levels (0..255; the bot clamps level 0 to 1 / 0, never offered), levels >= 1: merc.ReviveCost differs in {bad_m}, resource.MercReviveCost in {bad_r}")


main()
