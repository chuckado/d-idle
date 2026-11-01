#!/bin/sh
# SKABRECHECK driver (worker host). usage: skab_run.sh WHICH SEED   (WHICH = resist | phys)
# game.exe side under the emulator (about 2 minutes), then the d2d side from the fork build, then the compare.
W=$(cd "$(dirname "$0")/../.." && pwd)
. "$W/tools/emu/skab_env.sh"
B=${SKAB_BUILD:-$HOME/d2d-lab/fork/build-fr2/tests}
O=${SKAB_OUT:-/tmp/skab}
mkdir -p "$O"
cd "$D2/tools/emu" || exit 1
which=$1; seed=${2:-7}
case $which in
resist)
	$PY "$W/tools/emu/sk_resist.py" gen 100000 "$seed" > "$O/r$seed.cases"
	$PY "$W/tools/emu/sk_resist.py" game "$O/r$seed.cases" > "$O/r$seed.game"
	"$B/test_skab_resist" --resist "$O/r$seed.cases" | grep -E '^[0-9]+ ' > "$O/r$seed.d2d"
	$PY "$W/tools/emu/sk_resist.py" cmp "$O/r$seed.cases" "$O/r$seed.game" "$O/r$seed.d2d"
	$PY "$W/tools/emu/skab_pierce.py" "$O/r$seed.cases" "$O/r$seed.game" "$O/r$seed.d2d"
	;;
phys)
	$PY "$W/tools/emu/skab_phys.py" gen 100000 "$seed" > "$O/p$seed.cases"
	$PY "$W/tools/emu/skab_phys.py" game "$O/p$seed.cases" > "$O/p$seed.game"
	"$B/test_skab_phys" --phys "$O/p$seed.cases" "$O/p$seed.game" | grep -E '^[0-9]+ ' > "$O/p$seed.d2d"
	$PY "$W/tools/emu/skab_phys.py" cmp "$O/p$seed.cases" "$O/p$seed.game" "$O/p$seed.d2d"
	;;
esac
