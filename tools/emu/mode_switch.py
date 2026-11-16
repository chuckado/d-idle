# SPDX-License-Identifier: GPL-3.0-or-later
"""game.exe 1.14d: may a player unit enter get-hit (4) or block (9) from its current mode?  The oracle for
"a hit never restarts a get-hit" (d2d Fight::hurt_side, fork branch fid/hitrec-norestart).

    set -o pipefail; ../../build/tests/test_hit_recovery --hit-modes | grep '^M ' > d2d.tsv
    $PY mode_switch.py game > game.tsv         # game.exe's FUN_0057edd0 under unicorn (seconds)
    $PY mode_switch.py compare game.tsv d2d.tsv

`game` runs FUN_0057edd0 (may unit EAX enter mode EBX; ECX = the game; called by FUN_005809d0, the server's player mode switch)
over class 3 (Paladin; classes 0..6 are run too and a cell whose answer is class-dependent is printed as `C cls cur new queued allowed`),
current mode 0..18, new mode 4 and 9, with (queued = 1) and without (queued = 0) a pending type-1 mode-end event.
Hooks as hr_modeswitch.py (decomp20): FUN_005415a0 (the unit's queued type-1 event frame) returns the case's value, unit +0x60 = 0
(so FUN_0063c1e0 is not reached; a unit with a non-null +0x60 object may enter only mode 0x11 -- not run, U for players).
Output line: `M cur new queued allowed`.

`compare`: d2d's `M mode after restarted` (one big hit, standing in `mode` with a hold running, i.e. an event pending = queued 1).
d2d entered GH when restarted == 1 or (after == 4 and mode != 4); game.exe allowed is its queued=1 answer for new mode 4.
Cells the d2d rule leaves to other claims are listed and excluded by name, never silently: the swing modes A1 KK S1 S3 S4 (d2d never
interrupts a swing: card P04) and the dead modes DT DD (a dead unit)."""
import os
import sys

NAMES = "DT NU WL RN GH TN TW A1 A2 BL SC TH KK S1 S2 S3 S4 DD KB".split()
EXCLUDED = {"DT": "dead unit", "DD": "dead unit", "A1": "swing, P04", "KK": "swing, P04", "S1": "swing, P04", "S3": "swing, P04", "S4": "swing, P04"}
PALADIN = 3
FRAME = 5000


def game():
    sys.path.insert(0, os.environ["EMU"])
    import emu
    e = emu.Emu(os.environ["GAMEEXE"])
    state = {"end": 0}
    e.hook(0x5415a0, lambda e_: state["end"], 1)
    unit = e.alloc(0x200)
    gm = e.alloc(0x2000)
    e.w32(gm + 0xa8, FRAME)
    out = {}
    for cls in range(7):
        for queued in (0, 1):
            state["end"] = FRAME + 40 if queued else 0
            for cur in range(0, 19):
                for new in (4, 9):
                    e.w32(unit, 0)
                    e.w32(unit + 4, cls)
                    e.w32(unit + 0x10, cur)
                    e.w32(unit + 0x60, 0)
                    out[(cls, queued, cur, new)] = e.call(0x57edd0, ecx=gm, regs={"eax": unit, "ebx": new}) & 0xff != 0
    n = 0
    for (cls, queued, cur, new), v in sorted(out.items()):
        if cls == PALADIN:
            print("M %d %d %d %d" % (cur, new, queued, int(v)))
            n += 1
        elif v != out[(PALADIN, queued, cur, new)]:
            print("C %d %d %d %d %d" % (cls, cur, new, queued, int(v)))
    if n == 0:
        sys.exit(1)


def compare(gpath, dpath):
    allowed = {}
    for line in open(gpath):
        f = line.split()
        if len(f) == 5 and f[0] == "M":
            allowed[(int(f[1]), int(f[2]), int(f[3]))] = int(f[4])
    n = bad = 0
    for line in open(dpath):
        f = line.split()
        if len(f) != 4 or f[0] != "M":
            continue
        mode, after, restarted = int(f[1]), int(f[2]), int(f[3])
        name = NAMES[mode]
        entered = restarted == 1 or (after == 4 and mode != 4)
        g = allowed[(mode, 4, 1)]
        if name in EXCLUDED:
            print("excluded %s (%s): d2d entered %d, game allowed %d" % (name, EXCLUDED[name], int(entered), g))
            continue
        n += 1
        if int(entered) != g:
            bad += 1
            print("DIFF mode %s: d2d entered GH %d (after %s, restarted %d), game.exe allowed %d" % (name, int(entered), NAMES[after], restarted, g))
    print("%d cells, %d differ" % (n, bad))
    sys.exit(1 if bad or n == 0 else 0)


if __name__ == "__main__":
    if len(sys.argv) >= 2 and sys.argv[1] == "game":
        game()
    elif len(sys.argv) == 4 and sys.argv[1] == "compare":
        compare(sys.argv[2], sys.argv[3])
    else:
        sys.exit(__doc__)
