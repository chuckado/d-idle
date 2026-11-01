# SPDX-License-Identifier: GPL-3.0-or-later
"""palskill: print the non-empty columns of Skills.txt / Missiles.txt rows by name.
    python3 pal_rows.py Skills.txt skill "Holy Bolt,Blessed Hammer"
    python3 pal_rows.py Missiles.txt Missile "holybolt,blessedhammer"
Skills.txt / Missiles.txt: mpq-cat output from the patched 1.14d stack (D)."""
import csv
import sys

f, col, keys = sys.argv[1], sys.argv[2], sys.argv[3]
rows = list(csv.reader(open(f, encoding="latin1"), delimiter="\t"))
h = rows[0]
want = [k.lower() for k in keys.split(",")]
for r in rows[1:]:
    if r and r[h.index(col)].lower() in want:
        print("-----")
        for i, v in enumerate(r):
            if i < len(h) and v not in ("", "0"):
                print(f"  {h[i]} = {v}")
