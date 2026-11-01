#!/bin/bash
# palskill (the worker host): compile tools/emu/pal_test.cpp by hand against the
# libraries of a scratch upstream build (default: the skmana scratch build, same
# upstream commit as the scratch checkout in ~/d2d-lab/pr/palskill).
#   tools/emu/pal_build.sh [BUILD_DIR] -> $BUILD_DIR/../palskill-out/pal_test
set -e
SRC=$HOME/d2d-lab/pr/palskill
B=${1:-$HOME/d2d-lab/pr/skmana-out/build}
OUT=$HOME/d2d-lab/pr/palskill-out
IN=""
for d in tests components/game components/game/character components/game/gamedata components/game/npc components/log components/mpq components/d2s components/txt components/rules components/drlg components/ds1 components/userdir components/install components/cof components/compcode components/dt1 components/tbl; do IN="$IN -I$SRC/$d"; done
cd "$B"
nice c++ -DD2D_TEST_SKIPPED=77 -DSTORMLIB_NO_AUTO_LINK $IN -I$HOME/d2d-lab/src/build/_deps/stormlib-src/src -O1 -std=c++26 -include $SRC/cmake/no_near_far.hpp -Wno-missing-field-initializers -o $OUT/pal_test $SRC/tools/emu/pal_test.cpp components/game/libgame.a _deps/stormlib-build/libstorm.a /usr/lib/libz.so /usr/lib/libbz2.so components/install/libinstall_detect.a
echo built $OUT/pal_test
