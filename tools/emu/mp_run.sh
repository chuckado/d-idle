#!/bin/bash
# monsterpace (worker host): game.exe's thinks (mai2_ais.py, which imports mai_ais.py and upstream's mon_ais.py) -> cases -> d2d's REAL
# rules::mon_think replayed over them with a per-AI tally of cases and differences (mp_patch.py cuts the replay block out of the d2d tree's
# test at build time, in the scratch dir; the tree is only read).
#   tools/emu/run-on-worker.sh --ttl 1500 mp_run.sh <ai,ai|all> <cases> <seed> [MAI2_CMD none1|two|-] [less]
# env: D2D_TREE (default ~/d2d-lab/fork/wt-monsterpace), D2_SCRATCH
set -e
here=$(cd "$(dirname "$0")" && pwd)
SRC=${D2D_TREE:-$HOME/d2d-lab/fork/wt-monsterpace}
S=${D2_SCRATCH:-/tmp/d2agent-monsterpace}
AIS=${1:-all}; N=${2:-1500}; SEED=${3:-1}; CMD=${4:--}
[ "$CMD" = "-" ] && CMD=
LESS=${5:-}
UP=$SRC
PY=$HOME/d2d-lab/pr/walk2-venv/bin/python
mkdir -p "$S/run"
if [ "$AIS" = "all" ]; then AIS=Zombie,Brute,BloodHawk,CorruptRogue,CorruptLancer,Fallen,BloodRaven; fi
( cd "$UP/tools/emu" && MAI2_CMD=$CMD PYTHONPATH=$here:$UP/tools/emu D2_MPQ_DIR=/tmp/ug-emu nice "$PY" "$here/mai2_ais.py" "$AIS" "$N" "$SEED" --dump "$S/run" ) > "$S/run/acts.txt" 2>&1 || { tail -20 "$S/run/acts.txt"; exit 1; }
tail -1 "$S/run/acts.txt"
mv "$S/run/mai2_cases.inc" "$S/run/mp_cases.inc"; mv "$S/run/mai2_names.inc" "$S/run/mp_names.inc"
python3 "$here/mp_patch.py" "$SRC" "$S/run"
IN="-I$S/run"
for d in tests components/game components/game/character components/game/gamedata components/game/npc components/log components/mpq components/d2s components/txt components/rules components/drlg components/ds1 components/userdir components/install components/cof components/compcode components/dt1 components/tbl; do IN="$IN -I$SRC/$d"; done
IN="$IN -I$HOME/d2d-lab/src/build/_deps/stormlib-src/src"
nice c++ -std=c++26 -O1 -include "$SRC/cmake/no_near_far.hpp" -Wno-missing-field-initializers -Wno-error $IN -o "$S/run/mp_replay" "$S/run/mp_replay.cpp"
cd "$SRC"
set +e
"$S/run/mp_replay" $LESS 2> "$S/run/stderr.txt" | grep '^mp'
echo "stderr lines: $(wc -l < "$S/run/stderr.txt")"
