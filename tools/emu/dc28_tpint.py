# SPDX-License-Identifier: GPL-3.0-or-later
"""decomp28: what may enter a new mode while a player is casting / swinging a given skill?  game.exe 1.14d, FUN_0057edd0
(may the unit enter the mode) and FUN_0057eec0 (the used-skill gate), the pair FUN_005809d0 / FUN_00580a70 call for a hit,
a block, a knock-back and a command.  The extension of decomp26's interrupt_swing.py (that one only asked new mode 4).

    $PY dc28_tpint.py game > game.tsv      # seconds, one core

Cases: used skill (Attack 0, Fire Bolt 36, Teleport 54, Smite 97, Zeal 106, TP scroll 219, TP tome 220, ID scroll 217),
current mode (NU WL RN SC A1 TH), new mode 0..19, unit state (none, 0x36 uninterruptable, 0xf concentrate, 0x2a concentration with
the roll below / above its stat), a mode-end event queued (a cast has one) or not, the event already past (more than 5 frames) or not;
class 3 (Paladin) plus the other six classes (a class-dependent cell is printed as `C`).
Output `G skill state evt cur new allowed` (allowed = FUN_0057edd0 and FUN_0057eec0(new, 1, 0) both yes)."""
import os, sys
sys.path.insert(0, os.environ["EMU"])
import skill_tips
from unicorn.x86_const import UC_X86_REG_ESP

SKILLS = {0: "Attack", 36: "FireBolt", 54: "Teleport", 97: "Smite", 106: "Zeal", 217: "IDscroll", 219: "TPscroll", 220: "TPtome"}
CUR = {1: "NU", 2: "WL", 3: "RN", 10: "SC", 7: "A1", 11: "TH"}
STATES = {"none": (), "u36": (0x36,), "conc15": (0xf,), "conc2a_lo": (0x2a,), "conc2a_hi": (0x2a,)}
EVT = {"live": 40, "none": None, "past": -20}
FRAME = 5000
PALADIN = 3


def main():
    e = skill_tips.boot()
    cur = {"id": 0, "end": 0, "states": (), "roll": 0}
    e.hook(0x620250, lambda e_: 0xA0000000, 1)
    e.hook(0x643ce0, lambda e_: cur["id"], 3)
    def hasstate(e_):
        sp = e_.mu.reg_read(UC_X86_REG_ESP)
        return int(e_.r32(sp + 8) in cur["states"])
    e.hook(0x639df0, hasstate, 2)
    e.hook(0x5415a0, lambda e_: cur["end"], 1)
    e.hook(0x6256b0, lambda e_: 0, 2)        # the unit's stat table (concentration chance read below)
    e.hook(0x625d00, lambda e_: 50, 3)       # stat 0xa4 (skill_concentration) value: 50 percent
    e.hook(0x45c390, lambda e_: cur["roll"], 0)   # the percentage roll (ECX / EDX convention, args in registers)
    unit = e.alloc(0x200)
    gm = e.alloc(0x2000)
    e.w32(gm + 0xa8, FRAME)
    out = {}
    for cls in range(7):
        for sid in SKILLS:
            cur["id"] = sid
            for sname, st in STATES.items():
                cur["states"] = st
                cur["roll"] = 10 if sname == "conc2a_lo" else 90
                for en, d in EVT.items():
                    cur["end"] = 0 if d is None else FRAME + d
                    for cm in CUR:
                        for nm in range(0, 20):
                            e.mu.mem_write(unit, b"\0" * 0x200)
                            e.w32(unit + 4, cls)
                            e.w32(unit + 0x10, cm)
                            e.w32(unit + 0x60, 0)
                            a = e.call(0x57edd0, ecx=gm, regs={"eax": unit, "ebx": nm}) & 0xff != 0
                            b = e.call(0x57eec0, nm, 1, 0, ecx=gm, edx=unit) & 0xff != 0
                            out[(cls, sid, sname, en, cm, nm)] = a and b
    n = 0
    for k, v in sorted(out.items()):
        cls, sid, sname, en, cm, nm = k
        if cls == PALADIN:
            print("G %d %s %s %d %d %d" % (sid, sname, en, cm, nm, int(v)))
            n += 1
        elif v != out[(PALADIN,) + k[1:]]:
            print("C %d %d %s %s %d %d %d" % (cls, sid, sname, en, cm, nm, int(v)))
    if n == 0:
        sys.exit(1)


if __name__ == "__main__":
    if sys.argv[1:] == ["game"]:
        main()
    else:
        sys.exit("usage: dc28_tpint.py game")
