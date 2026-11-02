"""monsterai2: which monstats rows have a Fallen / FallenShaman AI row as minion1 / minion2 (D8: who could send a command to a Fallen).
usage: python3 mai2_minions.py <monstats table dump from mpq-table.sh monstats>"""
import csv
import sys

lines = open(sys.argv[1]).read().split("\n")
start = next(k for k, line in enumerate(lines) if line.startswith("Id\thcIdx"))
rows = list(csv.DictReader(lines[start:], delimiter="\t"))
ai = {r["Id"]: r["AI"] for r in rows}
fallen = {k for k, a in ai.items() if a in ("Fallen", "FallenShaman")}
print("fallen-AI rows:", sorted(fallen))
for r in rows:
    for m in ("minion1", "minion2"):
        if r[m] in fallen:
            print(f'leader {r["Id"]} (AI {r["AI"]}) {m} = {r[m]} (AI {ai[r[m]]})')
