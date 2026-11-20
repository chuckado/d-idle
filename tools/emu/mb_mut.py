#!/usr/bin/env python3
"""decomp43: the mutant runner for the monster-block claim (builder, long lane). One text change per mutant on a copy of the claim's
tree (/data/src/mbm), rebuild test_monster_block in /data/build/mbm, run its pins (no argument) and the compare against game.exe's
2,469 cases. Edited files are touched, and restored from the pristine copy after each mutant (an old mtime on a restored header
leaves stale objects, decomp42 trap).

    mb_mut.py SRC_MIRROR              (the claim's checkout on the builder)
"""
import os
import shutil
import subprocess
import sys

SRC = sys.argv[1]
W, B, R = "/data/src/mbm", "/data/build/mbm", "/data/d43/r"
ENV = dict(os.environ, D2_MPQ_DIR="/data/prfix6-emu/gd", D2_PATCH_INSTALLER="/data/p565-emu/LODPatch_114d.exe", D2_BUILD="/data/build/cr3",
           PYTHONPATH="/data/d43/tools/emu:" + SRC + "/tools/emu")
GATE = "if (!type_info.no_shield_block && !game_data.shield_code(npc.comp[7])) return 0;"
RET = "return std::min(type_info.diff[std::size_t(difficulty)].to_block, 75);"
MUTANTS = [
    ("m1 the parent rule (mBL)", "components/game/ai.hpp", GATE, "if (!type_info.can_block) return 0;"),
    ("m2 gate A dropped", "components/game/ai.hpp", GATE, "if (!game_data.shield_code(npc.comp[7])) return 0;"),
    ("m3 gate B dropped", "components/game/ai.hpp", GATE, "if (!type_info.no_shield_block) return 0;"),
    ("m4 no gate at all", "components/game/ai.hpp", GATE, ""),
    ("m5 cap removed", "components/game/ai.hpp", RET, "return type_info.diff[std::size_t(difficulty)].to_block;"),
    ("m6 cap 74", "components/game/ai.hpp", RET, RET.replace("75", "74")),
    ("m7 difficulty ignored", "components/game/ai.hpp", RET, RET.replace("std::size_t(difficulty)", "0")),
    ("m8 layer 6 for layer 7", "components/game/ai.hpp", GATE, GATE.replace("comp[7]", "comp[6]")),
    ("m9 mBL ANDed", "components/game/ai.hpp", GATE, GATE.replace("if (!type", "if (!type_info.can_block || (!type").replace("return 0;", ")) return 0;").replace("comp[7]))", "comp[7]))")),
    ("m10 shield by component column", "components/game/gamedata/gamedata.hpp", 'piece->second.type == "shie"', "piece->second.component == 7"),
    ("m11 any item code is a shield", "components/game/gamedata/gamedata.hpp", 'piece != item_pieces.end() && piece->second.type == "shie"', "piece != item_pieces.end()"),
    ("m12 gate A from the boss column", "components/game/gamedata/gamedata_load.cpp", 'text("NoShldBlock") == "1"', 'text("boss") == "1"'),
    ("m13 shie spelt shld (Any Shield)", "components/game/gamedata/gamedata.hpp", '"shie"', '"shld"'),
    ("m14 NoShldBlock any non-empty", "components/game/gamedata/gamedata_load.cpp", 'text("NoShldBlock") == "1"', 'text("NoShldBlock") != ""'),
]
only = sys.argv[2:] 


def sh(cmd, **kw):
    return subprocess.run(cmd, shell=True, capture_output=True, text=True, env=ENV, **kw)


def main():
    if not os.path.isdir(W):
        shutil.copytree(SRC, W, symlinks=True, ignore=shutil.ignore_patterns(".git"))
    else:
        sh(f"rsync -a --delete --exclude .git {SRC}/ {W}/")
    D = "/data/deps"
    if not os.path.exists(B + "/build.ninja"):
        r = sh(f"cmake -S {W} -B {B} -G Ninja -DCMAKE_BUILD_TYPE=Release -DD2D_WERROR=OFF -DCMAKE_C_COMPILER_LAUNCHER=ccache -DCMAKE_CXX_COMPILER_LAUNCHER=ccache -DD2D_BUNDLED_STORMLIB=OFF -DFETCHCONTENT_SOURCE_DIR_CLI11={D}/cli11-src -DFETCHCONTENT_SOURCE_DIR_SENTRY_NATIVE={D}/sentry_native-src -DFETCHCONTENT_SOURCE_DIR_STORMLIB={D}/stormlib-src")
        print("configure rc", r.returncode, flush=True)
    for name, path, old, new in MUTANTS:
        if only and name.split()[0] not in only:
            continue
        sh(f"rsync -a --delete --exclude .git {SRC}/ {W}/")
        f = os.path.join(W, path)
        s = open(f).read()
        if s.count(old) != 1:
            print(name, "PATTERN NOT FOUND", s.count(old), flush=True)
            continue
        open(f, "w").write(s.replace(old, new))
        sh(f"touch {f}")
        r = sh(f"nice cmake --build {B} -j6 --target test_monster_block")
        if r.returncode:
            print(name, "BUILD FAILED", r.stdout[-300:], flush=True)
            continue
        b = f"{B}/tests/test_monster_block"
        pins = sh(b + " > /dev/null 2>&1")
        d2d = sh(f"{b} --monster-block {R}/cases.tsv 2>/dev/null | grep -E '^[0-9]+ -?[0-9]+$' > {R}/mut.tsv; python3 {SRC}/tools/emu/monster_block.py compare {R}/cases.tsv {R}/game.tsv {R}/mut.tsv | tail -1")
        print(name, "| pins rc", pins.returncode, "(abort = nonzero) |", d2d.stdout.strip(), flush=True)
    sh(f"rsync -a --delete --exclude .git {SRC}/ {W}/")


main()
