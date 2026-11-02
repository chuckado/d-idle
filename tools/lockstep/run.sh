#!/bin/bash
# Run the P0 probes under Wine (console only, no display). Run on the worker host:
#   labrun orion -- tools/lockstep/run.sh [all|one MODE KPCT [SLEEP [MODS]]]
# Scratch Wine prefix: $LOCKSTEP_PREFIX (default ~/lockstepp0-scratch/prefix). Never a game prefix.
set -eu
cd "$(dirname "$0")"
OUT=${OUT:-build}
export WINEPREFIX=${LOCKSTEP_PREFIX:-$HOME/lockstepp0-scratch/prefix}
export WINEDEBUG=-all WINEDLLOVERRIDES=winemenubuilder.exe=d
unset DISPLAY WAYLAND_DISPLAY
mkdir -p "$(dirname "$WINEPREFIX")"
[ -d "$WINEPREFIX" ] || wine wineboot -i >/dev/null 2>&1 || true

# one MODE KPCT [SLEEP [MODS]]: KPCT "off" = hook DLL present but not patching
one() {
    mode=$1; k=$2; sl=${3:-0}; mods=${4-}
    unset LOCKSTEP_KPCT LOCKSTEP_SLEEP LOCKSTEP_MODS
    [ "$k" = off ] || export LOCKSTEP_KPCT=$k LOCKSTEP_SLEEP=$sl
    [ -z "$mods" ] || export LOCKSTEP_MODS=$mods
    export PROBE_MODE=$mode PROBE_SECS=${PROBE_SECS:-10}
    t0=$(date +%s.%N)
    ( cd "$OUT"; TIMEFORMAT="HOSTCPU user=%U sys=%S real=%R"; time wine probe.exe ) 2>&1 | tr -d '\r'
    t1=$(date +%s.%N)
    echo "HOSTWALL $(echo "$t1 $t0" | awk '{printf "%.3f", $1-$2}')"
}

if [ "${1:-all}" = one ]; then shift; one "$@"; exit 0; fi

PROBE_SECS=${PROBE_SECS:-10}
echo "== sleep granularity (unhooked)"; one sleep off
echo "== clocks 20 s unhooked"; PROBE_SECS=20 one clocks off
for k in 100 200 400; do
    echo "== clocks k=$k pct, sleep raw"; one clocks $k 0
done
echo "== clocks k=400 pct, sleep scaled"; one clocks 400 1
for k in off 100 200 400; do
    echo "== gate k=$k, raw sleep"; one gate $k 0
done
echo "== gate k=400, scaled sleep"; one gate 400 1
echo "== gate k=400, exe only patched (gamedll.dll not patched)"; one gate 400 0 exe
