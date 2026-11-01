#!/bin/bash
# necskill (the worker host): compile tools/emu/nec_test.cpp by hand against the
# libraries of a scratch upstream build (default: the skmana scratch build, same
# upstream commit as the read checkout ~/d2d-lab/pr/palskill, whose headers it uses).
#   tools/emu/nec_build.sh [BUILD_DIR] -> ~/d2d-lab/pr/necskill-out/nec_test
# Run from the worktree (the source is next to this script).
set -e
mkdir -p $HOME/d2d-lab/pr/necskill-out
SRC=$HOME/d2d-lab/pr/palskill
HERE=$(cd "$(dirname "$0")" && pwd)
B=${1:-$HOME/d2d-lab/pr/skmana-out/build}
OUT=$HOME/d2d-lab/pr/necskill-out
IN=""
for d in tests components/game components/game/character components/game/gamedata components/game/npc components/log components/mpq components/d2s components/txt components/rules components/drlg components/ds1 components/userdir components/install components/cof components/compcode components/dt1 components/tbl; do IN="$IN -I$SRC/$d"; done
cd "$B"
nice c++ -DD2D_TEST_SKIPPED=77 -DSTORMLIB_NO_AUTO_LINK $IN -I$HOME/d2d-lab/src/build/_deps/stormlib-src/src -O1 -std=c++26 -include $SRC/cmake/no_near_far.hpp -Wno-missing-field-initializers -o $OUT/nec_test $HERE/nec_test.cpp components/game/libgame.a _deps/stormlib-build/libstorm.a /usr/lib/libz.so /usr/lib/libbz2.so components/install/libinstall_detect.a
echo built $OUT/nec_test
