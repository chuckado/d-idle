#!/bin/sh
# SKABRECHECK claim 2, before / after (worker host). Standalone compile of the d2d test against a rules header:
# the physical roll test needs only the header-only rules code (combat.hpp, rules.hpp), no game library.
#   skab_fix.sh before|after
# before = the lab binary's source (fork lab2-freezeresume, ee5e06d8); after = the same plus the local branch
# local/skabrecheck-phys (one hunk in components/rules/combat.hpp). Needs /tmp/skab/p7.{cases,game}.
set -e
W=$(cd "$(dirname "$0")/../.." && pwd)
. "$W/tools/emu/skab_env.sh"
F=$HOME/d2d-lab/fork
O=${SKAB_OUT:-/tmp/skab}
case $1 in before) R=$F/wt-lab2s ;; after) R=$F/wt-skabrecheck ;; *) echo usage; exit 2 ;; esac
inc=""
for d in tests components/game components/game/character components/game/gamedata components/game/npc components/log components/mpq components/d2s components/txt components/rules components/drlg components/ds1 components/userdir components/install components/cof components/compcode components/dt1 components/tbl; do inc="$inc -I$R/$d"; done
c++ -DD2D_TEST_SKIPPED=77 -DSTORMLIB_NO_AUTO_LINK $inc -I$F/build-fr2/_deps/stormlib-src/src -O3 -std=c++26 -include $R/cmake/no_near_far.hpp -Wno-missing-field-initializers \
	-o "$O/skab_phys_$1" "$W/tools/emu/skab_phys_test.cpp"
"$O/skab_phys_$1" --phys "$O/p7.cases" "$O/p7.game" | grep -E '^[0-9]+ ' > "$O/p7.d2d.$1"
cd "$D2/tools/emu"
$PY "$W/tools/emu/skab_phys.py" cmp "$O/p7.cases" "$O/p7.game" "$O/p7.d2d.$1"
