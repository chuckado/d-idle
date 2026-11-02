# SPDX-License-Identifier: GPL-3.0-or-later
"""mercstats: game.exe's merc think (FUN_005e52d0) as the oracle for d2d's C++ rules::hireable_think.

    mst_think.py CASES SEED     one line per case:
        T cls size merc_mode mx my ox oy owner_mode ex ey cursor on40 town target split lo hi BITS rx0 ry0 .. rx19 ry19 ::: effects ::: lo hi

The setups and the hooks are upstream's tools/emu/merc.py (`setup`, `install`, `run_game`); the effects are what game.exe's
own code did (moves set off with the path ctx, a stand, a teleport, the attack hand-off) and the merc's seed after.
BITS: 100 characters, the n-th move that is tried finds a path when it is `1` (the same random sequence the game run used).
tools/emu/mst_d2d_test.cpp `think` reads these lines, runs d2d's real rules::hireable_think on the same setup and prints the
line again with its own effects; the two files must be equal.
"""
import random
import sys

import emu
import merc as think


def fx(items):
    parts = []
    for it in items:
        if it[0] == "move":
            parts.append("move %d %d %d %d %d %d %d" % it[1:8])
        elif it[0] == "stand":
            parts.append("stand %d" % it[1])
        else:
            parts.append(it[0])
    return ";".join(parts)


def main():
    cases, seed0 = int(sys.argv[1]), int(sys.argv[2])
    if len(sys.argv) > 3:                       # an output file (under $D2_SCRATCH on the worker host) instead of stdout
        sys.stdout = open(sys.argv[3], "w")
    rng = random.Random(seed0)
    e = emu.Emu()
    think.install(e)
    mark = e.brk
    for _ in range(cases):
        e.brk = mark
        s = think.setup(rng)
        move_seed = rng.getrandbits(32)
        rm = random.Random(move_seed)
        bits = "".join("1" if rm.random() >= s["fails"] else "0" for _ in range(100))
        got, seed_after = think.run_game(e, s, random.Random(move_seed))
        split = s["split"] if s["split"] is not None else 99999999
        ring = " ".join("%d %d" % p for p in s["ring"])
        target = -1 if s["target"] is None else s["target"]
        print("T %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %d %s %s ::: %s ::: %d %d" % (
            s["cls"], s["size"], s["merc_mode"], s["merc"][0], s["merc"][1], s["owner"][0], s["owner"][1], s["owner_mode"],
            s["end"][0], s["end"][1], s["cursor"], int(s["on40"]), int(s["town"]), target, split, s["seed"][0], s["seed"][1],
            bits, ring, fx(got), seed_after[0], seed_after[1]))


main()
