# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp31: game.exe's roll against d2d's rule (drop-dump of the pristine fork build) on the exact (class, ilvl) keys of the Act 1 Normal population.

  bash tools/emu/tre_run.sh dm_diff.py POP.txt [SEEDS_PER_KEY]
Keys are the same as dm_model.py builds (unit class + flags -> treasure class, ilvl = stat 12), plus the quest class of every boss with a TCQuestId.
Each key is rolled SEEDS times with one player and magic find 0 and also at 3 players and magic find 300; the lines of game.exe
(drops.Oracle.roll) and of d2d (drop-dump) must be equal.  Exit 1 on any difference.
"""
import sys
from collections import defaultdict

import drops
import tre_diff  # noqa: F401  (points drops.run_ours at $TRE_DROPDUMP)

POP = sys.argv[1]
SEEDS = int(sys.argv[2]) if len(sys.argv) > 2 else 300


def main():
    o = drops.Oracle()
    e = o.e
    ms = {r["hcIdx"]: r for r in drops.txt_rows(e, "MonStats") if r.get("Id") and r.get("hcIdx", "").isdigit()}
    sup = defaultdict(list)
    for r in drops.txt_rows(e, "SuperUniques"):
        if r.get("Superunique") and r["Superunique"] != "Expansion":
            sup[r["Class"].lower()].append(r)
    keys = set()
    for line in open(POP):
        f = line.split()
        if not f or f[0] != "U" or f[6] == "None":
            continue
        cls, flags = int(f[4]), int(f[5], 16)
        st = {}
        for tok in f[8:]:
            k, _, v = tok.partition(":")
            try: st[int(k, 16)] = int(v)
            except ValueError: pass
        if st.get(0xd, 0) <= 0:
            continue
        r = ms[str(cls)]
        mid = r["Id"].lower()
        if flags & 2 and len(sup.get(mid, [])) == 1: tc = sup[mid][0]["TC"]
        elif flags & 2: tc = r["TreasureClass3"]
        elif flags & 4: tc = r["TreasureClass2"]
        elif flags & 8: tc = r["TreasureClass3"]
        else: tc = r["TreasureClass1"]
        il = max(st.get(0xc, int(f[6])), 1)
        if tc in o.index:
            keys.add((tc, il))
        if r.get("TCQuestId") not in ("", "0", None) and r.get("TreasureClass4") in o.index:
            keys.add((r["TreasureClass4"], il))
    keys = sorted(keys)
    todo = []
    for n, (tc, il) in enumerate(keys):
        for s in range(SEEDS):
            seed = ((n * 100003 + s + 1) * 0x9E3779B1) & 0xFFFFFFFF
            todo.append((seed, tc, 0, il, 1, 0))
            if s % 3 == 0:
                todo.append((seed ^ 0x5555, tc, 0, il, 3, 300))
    game = [o.roll(*j) for j in todo]
    ours = drops.run_ours([], "".join("\t".join(map(str, j)) + "\n" for j in todo))
    bad = [(g, u, j) for g, u, j in zip(game, ours, todo) if g != u and j[1] not in drops.KNOWN]
    for g, u, j in bad[:8]:
        print(f"game: {g}\nours: {u}")
    per = defaultdict(lambda: [0, 0])
    for g, u, j in zip(game, ours, todo):
        per[j[1]][0] += 1
        per[j[1]][1] += g != u
    print(f"keys {len(keys)}, classes {len(per)}, rolls {len(todo)}, equal {len(todo) - len(bad)}, differ {len(bad)}")
    sys.exit(1 if bad or len(game) != len(ours) else 0)


if __name__ == "__main__":
    main()
