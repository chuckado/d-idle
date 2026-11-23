#!/usr/bin/env python3
"""decomp43: the cheap mutant runner for the monster-block claim (builder). mb_mut.py rebuilds the whole game library per mutant
(ai.hpp reaches every translation unit: minutes each, GBs of memory). The rule lives in two inline functions of headers
(Monster::block_chance in ai.hpp, GameData::shield_code in gamedata.hpp) that test_monster_block's own translation unit inlines at its
only call, so this runner compiles just that unit against a shadow copy of the mutated header (first on the include path), links
it with the claim's already built libraries and runs the pins and the compare. The loader's NoShldBlock read lives in a library
unit and is NOT reachable this way (mb_mut.py m12 and m14 cover it).

    mb_mut2.py SRC BUILD [m5 m6 ...]       SRC the claim's tree, BUILD its build dir (test_monster_block built there)
"""
import os
import re
import shutil
import subprocess
import sys

SRC, B = sys.argv[1], sys.argv[2]
only = sys.argv[3:]
OV, R = "/data/src/mbm-ov", "/data/d43/r"
ENV = dict(os.environ, D2_MPQ_DIR="/data/prfix6-emu/gd", D2_PATCH_INSTALLER="/data/p565-emu/LODPatch_114d.exe", D2_BUILD="/data/build/cr3",
           PYTHONPATH="/data/d43/tools/emu:" + SRC + "/tools/emu")
GATE = "if (!type_info.no_shield_block && !game_data.shield_code(npc.comp[7])) return 0;"
RET = "return std::min(type_info.diff[std::size_t(difficulty)].to_block, 75);"
AI, GD = "components/game/ai.hpp", "components/game/gamedata/gamedata.hpp"
MUTANTS = [
    ("m5 cap removed", AI, RET, "return type_info.diff[std::size_t(difficulty)].to_block;"),
    ("m6 cap 74", AI, RET, RET.replace("75", "74")),
    ("m7 difficulty ignored", AI, RET, RET.replace("std::size_t(difficulty)", "0")),
    ("m8 layer 6 for layer 7", AI, GATE, GATE.replace("comp[7]", "comp[6]")),
    ("m9 mBL ANDed", AI, GATE, "if (!type_info.can_block || (!type_info.no_shield_block && !game_data.shield_code(npc.comp[7]))) return 0;"),
    ("m10 shield by component column", GD, 'piece->second.type == "shie"', "piece->second.component == 7"),
    ("m11 any item code is a shield", GD, 'piece != item_pieces.end() && piece->second.type == "shie"', "piece != item_pieces.end()"),
    ("m13 Any Shield (shld) for shie", GD, '"shie"', '"shld"'),
    ("m15 gate A only through can_block", AI, GATE, "if (!(type_info.no_shield_block && type_info.can_block) && !game_data.shield_code(npc.comp[7])) return 0;"),
]


def sh(cmd, **kw):
    return subprocess.run(cmd, shell=True, capture_output=True, text=True, env=ENV, **kw)


def main():
    cmds = sh(f"ninja -C {B} -t commands test_monster_block").stdout.splitlines()
    compile_cmd = next(c for c in cmds if "test_monster_block.cpp.o" in c and " -c " in c)
    link_cmd = next(c for c in cmds if " -o tests/test_monster_block" in c)
    obj = re.search(r"-o (\S*test_monster_block\.cpp\.o)", compile_cmd).group(1)
    print("compile unit", obj, flush=True)
    for name, path, old, new in MUTANTS:
        if only and name.split()[0] not in only:
            continue
        shutil.rmtree(OV, ignore_errors=True)
        os.makedirs(os.path.join(OV, "x"), exist_ok=True)
        text = open(os.path.join(SRC, path)).read()
        if text.count(old) != 1:
            print(name, "PATTERN NOT FOUND", text.count(old), flush=True)
            continue
        open(os.path.join(OV, os.path.basename(path)), "w").write(text.replace(old, new))
        mut_obj, mut_bin = "/tmp/mbm_mut.o", "/tmp/mbm_mut_bin"
        c = re.sub(r"(\S*(?:c\+\+|g\+\+)) ", lambda m: f"{m.group(1)} -I{OV} ", compile_cmd.replace(obj, mut_obj), count=1)
        r = sh(f"cd {B} && nice {c}")
        if r.returncode:
            print(name, "COMPILE FAILED", r.stderr[-300:], flush=True)
            continue
        l = link_cmd.replace(obj, mut_obj).replace(" -o tests/test_monster_block", f" -o {mut_bin}")
        r = sh(f"cd {B} && {l}")
        if r.returncode:
            print(name, "LINK FAILED", r.stderr[-300:], flush=True)
            continue
        pins = sh(mut_bin + " > /dev/null 2>&1")
        d2d = sh(f"{mut_bin} --monster-block {R}/cases.tsv 2>/dev/null | grep -E '^[0-9]+ -?[0-9]+$' > {R}/mut2.tsv; python3 {SRC}/tools/emu/monster_block.py compare {R}/cases.tsv {R}/game.tsv {R}/mut2.tsv | tail -1")
        print(name, "| pins rc", pins.returncode, "(abort = nonzero) |", d2d.stdout.strip(), flush=True)


main()
