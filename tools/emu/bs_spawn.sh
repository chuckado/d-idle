#!/bin/bash
# BOSS1 (decomp19): game.exe's own room population (umod_spawn.py) over the Act 1 Normal boss areas, difficulty 0.
# Levels.txt ids: 8 Den, 17 Burial Grounds, 25 Tower Cellar 5, 28 Barracks, 37 Catacombs 4, 38 Tristram, 3 Cold Plains,
# 4 Stony Field, 5 Dark Wood, 9 Cave 2. Output: $D2_SCRATCH/bs_a<area>.txt. Uses the emu image (docs/emu-image.md): env.sh.
set -u
. "${D2_EMU_DIR:-$HOME/d2d-lab/emu}/env.sh"
here=$(cd "$(dirname "$0")" && pwd)
# args: DIFF SEEDS AREA... (defaults 0, 1-4, 17 25 28 37 38 8); env does not pass through run-on-worker.sh
BS_DIFF=${1:-0}; BS_SEEDS=${2:-1-4}; shift 2 2>/dev/null; BS_AREAS=${*:-17 25 28 37 38 8}
out=${D2_SCRATCH:-/tmp}
export PYTHONPATH=$here:$EMU
cd "$EMU" || exit 2
for a in ${BS_AREAS:-17 25 28 37 38 8}; do
	nice "$PY" "$here/umod_spawn.py" ${BS_SEEDS:-1-4} $a ${BS_DIFF:-0} > "$out/bs_a${a}_d${BS_DIFF:-0}.txt" 2>&1
	echo "area=$a lines=$(wc -l < "$out/bs_a${a}_d${BS_DIFF:-0}.txt")"
done
