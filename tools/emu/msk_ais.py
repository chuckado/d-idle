"""monsterskill: the Corrupt Archer's think (MonAI row, 0x5f5a20: Dark Ranger / Vile Archer / Dark Archer / Black Archer / Flesh Archer, Act 1 and 2)
run natively and dumped for d2d's replay. Imports the monsterai2 harness (mai2_ais.py, and through it monsterai's mai_ais.py and upstream's mon_ais.py)
and only adds the think address; nothing else is changed.

    PYTHONPATH=<upstream tools/emu>:tools/emu D2_MPQ_DIR=<dir> python msk_ais.py CorruptArcher <cases> <seed> --dump <dir>
Writes <dir>/mai2_cases.inc and mai2_names.inc, which mai2_patch.py's replay program reads (tools/emu/msk_run.sh). Never runs a game."""
import sys

import mai_ais as A
import mai2_ais as B

import mon_ais as M

A.NEW["CorruptArcher"] = 0x5f5a20

_install = M.install
RAW_N = __import__("os").environ.get("MSK_RAWN", "1") == "1"    # 1: the step count as the think passed it (aip5); 0: with FUN_005defb0's own 0 -> 1 default


def install(e):
    """The Corrupt Archer runs in with FUN_005defb0 (move mode 0xf, RN: FUN_005deb60(unit, 0xf, 0, 0, 0, n, 0)), not FUN_005ded40 (mode 2 with
    FUN_005a6260(0xd)), which mon_ais hooks. Left unhooked, the real FUN_005defb0 reaches FUN_005dead0 and the harness labels the run a skill
    (the first run of this card: 551 of 1,500 'skill' for 'run'). Hooked here the same way as FUN_005ded40: ("run", target, n, 1), set off."""
    _install(e)
    def run_rn(e):
        cur = M.CUR
        cur["out"].append(("run", cur["names"].get(e.arg(0)), M.S32(e.arg(1)) if RAW_N else (M.S32(e.arg(1)) or 1), 1))
        return 1
    e.hook(0x5defb0, run_rn, 2)


M.install = install

if __name__ == "__main__":
    sys.exit(B.main())
