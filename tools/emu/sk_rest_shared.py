# SPDX-License-Identifier: GPL-3.0-or-later
"""SORCREST: does one Fire Ball / Glacial Spike burst roll its damage once for every unit?

    python sk_rest_shared.py [fireball|glacialspike [TRIALS]]    # prints `S trial rec1 rec2 rec3 same`

Runs next to upstream's tools/emu (missiles.py's Game). Three hostile monster
units stand within the burst radius; the missile (the real FUN_0059fa30) has
its element min / max forced through the stat getter (fire 0x30 / 0x31, cold
0x36 / 0x37: 1000 / 5000 in 256ths) and a different seed per trial; the burst
hit function runs natively (FUN_005a9a70 or FUN_005aa8b0, FUN_005a8c70 rolling
the record, FUN_0056bad0 handing it to each unit), FUN_0056b9c0 is hooked to
copy the 0x70-byte record each unit is given. `same` = 1 when the three
records are byte-equal (one roll for the burst); a roll per unit would differ
with the probability about 1 - 1/4000 per pair.
"""
import random
import sys

import missiles

ROW = {"fireball": 62, "glacialspike": 96}
SKILL = {"fireball": 47, "glacialspike": 55}
HITFN = {"fireball": 0x5a9a70, "glacialspike": 0x5aa8b0}
STATS = {"fireball": (0x30, 0x31), "glacialspike": (0x36, 0x37)}


def main():
    name = sys.argv[1] if len(sys.argv) > 1 else "fireball"
    trials = int(sys.argv[2]) if len(sys.argv) > 2 else 200
    game = missiles.Game()
    e = game.e
    rng = random.Random(7)
    lo, hi = STATS[name]
    forced = {lo: 1000, hi: 5000}
    foes, copies = set(), []
    e.hook(0x554200, lambda en: int(en.arg(0) in foes), 1)
    e.hook(0x625480, lambda en: forced.get(en.arg(1), 0) if en.arg(1) in forced else 0, 3)

    def on_apply(en):
        copies.append((en.arg(0), bytes(e.read(en.arg(1), 0x70))))
        return 1
    e.hook(0x56b9c0, on_apply, 3)
    spot = None
    for room in game.rooms:
        rx, ry, rw, rh = room[1:5]
        for _ in range(80):
            x, y = rx + rng.randrange(rw), ry + rng.randrange(rh)
            if all(game.collision(x + dx, y + dy) & 0x1f0f == 0 for dx in range(-8, 9, 2) for dy in range(-8, 9, 2)):
                spot = (x, y)
                break
        if spot: break
    x, y = spot
    units = [game.unit(19, x + dx, y + dy) for dx, dy in ((0, 0), (2, 1), (-1, 2))]
    shooter = game.unit(63, x + 12, y + 12) or game.unit(63, x - 12, y - 12)
    units = [u for u in units if u]
    foes.update(u[0] for u in units)
    print("# units", len(units), file=sys.stderr)
    same = 0
    for t in range(trials):
        p = e.alloc(0x60)
        for i, v in {0: 0x21, 1: shooter[0], 4: ROW[name], 5: x + 1, 6: y + 1, 7: x + 11, 8: y + 1, 0xb: SKILL[name], 0xc: 5}.items(): e.w32(p + 4 * i, v)
        m = e.call(0x59fa30, ecx=game.g, edx=p)
        e.w32(m + 0x20, rng.getrandbits(32))
        e.w32(m + 0x24, rng.getrandbits(32))
        del copies[:]
        e.call(HITFN[name], 0, ecx=game.g, edx=m)
        recs = [c[1] for c in copies]
        eq = int(len(recs) == len(units) and len(set(recs)) == 1)
        same += eq
        print("S", t, len(recs), " ".join(r[0x10:0x14].hex() + "/" + r[0x24:0x28].hex() for r in recs), eq)
    print(f"# {trials} trials, {same} with one shared record", file=sys.stderr)


if __name__ == "__main__":
    main()
