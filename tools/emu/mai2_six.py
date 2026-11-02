"""monsterai2: ports of the six thinks d2d's mon_think does not carry (Town Rogue, Elemental Beast, Flying Scimitar, Quill Mother,
Andariel, Dark Wanderer), written from the game.exe disassembly (docs/research/monsterai2-trace.md section 3) and checked against
game.exe's own code run under the emulator.

    PYTHONPATH=<upstream tools/emu>:tools/emu D2_MPQ_DIR=<dir> python mai2_six.py [cases] [seed] [--dump-andariel FILE]

A port is a function scene -> (what the think calls, in order, as upstream mon_ais.py's hooks log it; the three AI control words
after; the seed after). It is right when all three equal the game's, case by case. Imports upstream's mon_ais.py (scene, hooks) and
mai_ais.py (think addresses); edits neither. Never runs a game. Prints `six AI cases differ` lines."""
import random
import sys

import mai_ais as A
import mon_ais as M
from merc import Seed, spot_gap

AND_DUMP = []


def roller(s):
    sd = Seed(*s["seed"])
    return sd, lambda: sd.rand(100)


def skill_call(s, slot, unit="foe"):
    return ("skill", s["modes"][slot] & 0xff, s["skills"][slot], unit, 0, 0)


def town_rogue(s):
    # FUN_005ddc30: the second target and its distance; within 25 it strikes it (A1, mode 4), else stands 50.
    if s["foe2"] and s["foe2_dist"] < 25:
        return [("mode", 4, "foe2")], list(s["ctrl"]), s["seed"]
    return [("stand", 50)], list(s["ctrl"]), s["seed"]


def elemental_beast(s):
    ctrl, sd, roll = list(s["ctrl"]), *roller(s)
    aip = s["aip"]
    out = []
    if ctrl[0] == 2:
        return [("die",)], ctrl, (sd.lo, sd.hi)
    if ctrl[0] == 0 and (s["dist"] < aip[1] or s["got_hit"]):
        ctrl[0] = 1
        return [("mode", 8, "foe")], ctrl, (sd.lo, sd.hi)                    # S1 at the target, state 1
    if ctrl[0] == 1:
        if not s["melee"]:
            if roll() < aip[0]:
                return [("walk", "foe", 0, 1)], ctrl, (sd.lo, sd.hi)
            n = 8                                                           # FUN_005de200(8): the wander's draws
            return [("wander", n, *wander(sd, n))], ctrl, (sd.lo, sd.hi)
        ctrl[0] = 2
        return [("mode", 8, "foe"), ("event", 1)], ctrl, (sd.lo, sd.hi)
    return [("stand", aip[2])], ctrl, (sd.lo, sd.hi)


def wander(sd, n):                                                           # FUN_005de200 / mon_ais.wander_draws
    if sd.step() & 1 == 0: x, y = sd.rand(n), n
    else: x, y = n, sd.rand(n)
    if sd.step() & 1: x = -x
    if sd.step() & 1: y = -y
    return x, y


def circle(sd, unit, n, flag):                                               # FUN_005df7d0: one step, then a walk round it
    v = sd.step()
    return ("circle", unit, n, flag, int((v & 0xff) >= 0x80), 1)


def flying_scimitar(s):
    sd, roll = roller(s)
    aip = s["aip"]
    if s["melee"]:
        if roll() < aip[1]: out = [("mode", 4, "foe")]
        elif roll() < aip[3]: out = [circle(sd, "foe", 2, 0)]
        else: out = [("stand", aip[2])]
    else:
        if roll() < aip[0]: out = [("keep", "foe", 2, 8, 1)]                 # FUN_005de6d0(target, 8, 1) -> FUN_005de4e0(.., mode 2, 8, 1)
        elif roll() < aip[0]: out = [("walk", "foe", 7, 1)]
        else: out = [("stand", aip[2])]
    return out, list(s["ctrl"]), (sd.lo, sd.hi)


def quill_mother(s):
    sd, roll = roller(s)
    aip = s["aip"]
    if s["got_hit"]:
        out = [("rally",), ("mode", 4, "foe")] if s["melee"] else [("rally",), ("walk", "foe", 7, 1)]
    elif s["melee"]:
        out = [("mode", 4, "foe")] if roll() < aip[0] else [("stand", aip[2])]
    else:
        out = [("walk", "foe", 7, 1)] if roll() < aip[1] else [("stand", aip[3])]
    return out, list(s["ctrl"]), (sd.lo, sd.hi)


def andariel(s):
    sd, roll = roller(s)
    aip, sk = s["aip"], s["skills"]
    if s["melee"]:
        if sk[0] >= 0 and roll() < aip[0]: out = [skill_call(s, 0)]
        else: out = [("mode", 4, "foe")]
    elif roll() < aip[1]:
        out = [("stand", 5)]
    elif roll() < aip[2]:
        out = None
        if sk[0] >= 0 and roll() < aip[3]: out = [skill_call(s, 0)]
        elif sk[1] >= 0: out = [skill_call(s, 1)]
        if out is None: out = [("pace", 0, 0), ("walk", "foe", 7, 1)]
    else:
        out = [("pace", 0, 0), ("walk", "foe", 7, 1)]
    return out, list(s["ctrl"]), (sd.lo, sd.hi)


def dark_wanderer(s):
    sd = Seed(*s["seed"])
    ctrl = list(s["ctrl"])
    if not s["wfound"]:
        return [("stand", 10)], ctrl, s["seed"]
    if ctrl[0] == 0: ctrl = [1, 0, 0]
    if not s["target"]: return [("stand", 10)], ctrl, s["seed"]
    if s["dist"] >= 20: return [("stand", 40)], ctrl, s["seed"]
    x, y = s["wspot"]
    if ctrl[0] == 1:
        ctrl[0] = 2
        return [("walkto", x, y)], ctrl, s["seed"]
    if ctrl[0] == 2:
        mx, my = s["spots"]["mon"]
        if spot_gap(mx, my, x, y) < 2 or ctrl[1] >= 3:
            return [("gone", 0, 0, 0), ("quest",)], ctrl, s["seed"]
        ctrl[1] += 1
        return [("pace", 0, 0), ("walkto", x, y)], ctrl, s["seed"]
    return [("stand", 40)], ctrl, s["seed"]


PORTS = {"TownRogue": town_rogue, "ElementalBeast": elemental_beast, "FlyingScimitar": flying_scimitar, "QuillMother": quill_mother,
         "Andariel": andariel, "DarkWanderer": dark_wanderer}


def setup(rng, ai):
    s = M.setup(rng, ai)
    if ai == "DarkWanderer":                                                 # his spot (FUN_005bd0d0), near him often
        mx, my = s["spots"]["mon"]
        s["wfound"] = rng.random() < 0.85
        s["wspot"] = (mx + rng.randint(-4, 4), my + rng.randint(-4, 4)) if rng.random() < 0.6 else (mx + rng.randint(-40, 40), my + rng.randint(-40, 40))
        s["ctrl"][0] = rng.choice((0, 1, 2, 2, 3, rng.randint(0, 6)))
        s["ctrl"][1] = rng.choice((0, 1, 2, 3, 4))
    if ai == "ElementalBeast":
        s["ctrl"][0] = rng.choice((0, 0, 1, 1, 2, 3))
    if ai == "Andariel":
        s["skills"][0] = rng.choice((-1, 150, 150, 150, 150)); s["skills"][1] = rng.choice((-1, 151, 151, 151, 151))   # her row has both: mostly set
        s["skills"][2:] = [-1, -1, -1]
    return s


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    cases = int(args[0]) if args else 3000
    rng = random.Random(int(args[1]) if len(args) > 1 else 1)
    for name, addr in A.NEW.items(): M.AIS.setdefault(name, addr)
    M.ORDER[:] = list(M.AIS)
    e = M.emu.Emu()
    A.patch(e)
    M.install(e)
    e.hook(0x5bd0d0, lambda e: (e.w32(e.arg(0), M.CUR["s"]["wspot"][0]), e.w32(e.arg(0) + 4, M.CUR["s"]["wspot"][1]), int(M.CUR["s"]["wfound"]))[2], 1)
    e.hook(0x5bd4a0, lambda e: M.CUR["out"].append(("quest",)) or 0, 0)
    total = bad = 0
    for ai, port in PORTS.items():
        mark, n, differ, acts = e.brk, 0, 0, {}
        first = None
        for _ in range(cases):
            e.brk = mark
            s = setup(rng, ai)
            out, ctrl, seed = M.run_game(e, s, M.AIS[ai], random.Random(rng.getrandbits(32)))
            mine_out, mine_ctrl, mine_seed = port(s)
            act = M.outcome(s, out)[0][0]
            acts[act] = acts.get(act, 0) + 1
            same = [tuple(x) for x in out] == [tuple(x) for x in mine_out] and ctrl[:3] == mine_ctrl[:3] and tuple(seed) == tuple(mine_seed)
            n += 1
            if not same:
                differ += 1
                if first is None: first = (s, out, mine_out, ctrl, mine_ctrl)
            if ai == "Andariel" and "--dump-andariel" in sys.argv:
                AND_DUMP.append(" ".join(map(str, [s["diff"], int(s["melee"]), *s["aip"][:4], int(s["skills"][0] >= 0), int(s["skills"][1] >= 0), *s["seed"], M.ACTS.index(M.outcome(s, out)[0][0]), M.outcome(s, out)[0][1], *seed])))
        print(f"six {ai} {n} {differ} acts {dict(sorted(acts.items()))}")
        if first:
            s, out, mine_out, ctrl, mine_ctrl = first
            print("  first mismatch: game", out, ctrl, "port", mine_out, mine_ctrl, "dist", s["dist"], "melee", s["melee"], "aip", s["aip"], "skills", s["skills"])
        total += n; bad += differ
    print(f"six TOTAL {total} {bad}")
    if AND_DUMP:
        with open(sys.argv[sys.argv.index("--dump-andariel") + 1], "w") as f:
            f.write("\n".join(AND_DUMP) + "\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
