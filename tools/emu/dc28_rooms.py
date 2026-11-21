# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp28: the rooms game.exe's level generator builds for a level and seed (what the room population then fills).
    python dc28_rooms.py LO-HI LEVELIDS [difficulty]  ->  `S lid seed rooms tiles` (rooms of the level and the sum of room width x height at r+0x3c, r+0x40, in tiles; every room is 8 x 8, V run)"""
import os, sys
sys.path.insert(0, os.getcwd())
import drlg, monsters, objgroups


def rooms(e, seed, lid, difficulty):
    g = monsters.new_game(e, seed, difficulty)
    act = monsters.act_of(e, g, lid, drlg.act_of(lid))
    lvl = drlg.find_level(e, e.r32(act + 0x48), lid)
    if not lvl:
        raise SystemExit("level %d not built" % lid)
    out = []
    r = e.r32(lvl + 0x10)
    while r:
        out.append([e.s32(r + o) for o in range(0x30, 0x50, 4)])
        r = e.r32(r + 0x24)
    return out


def main():
    lo, _, hi = sys.argv[1].partition("-")
    lids = [int(x) for x in sys.argv[2].split(",")]
    diff = int(sys.argv[3]) if len(sys.argv) > 3 else 0
    e = drlg.boot()
    for lid in lids:
        for seed in range(int(lo), int(hi or lo) + 1):
            box = []
            try:
                objgroups.quest_stubbed(e, lambda e_, s, l, d: box.append(rooms(e_, s, l, d)) or "", seed, lid, diff)
            except Exception as err:
                print("# lid %d seed %d failed: %s" % (lid, seed, err))
                continue
            rms = box[0] if box else []
            # words: [0] 0, [1] x tiles, [2] y tiles, [3] width, [4] height (tiles; 8 x 8 in every room seen), [5..7] unknown
            print("S %d %d %d %d" % (lid, seed, len(rms), sum(r[3] * r[4] for r in rms)))


if __name__ == "__main__":
    main()
