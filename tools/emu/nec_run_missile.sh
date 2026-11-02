#!/bin/bash
# necskill: magic missile hit differential on the worker host.
#   tools/emu/nec_run_missile.sh N SEED [OUTDIR]
set -e
HERE=$(cd "$(dirname "$0")" && pwd)
N=${1:-20000}; SEED=${2:-1}; OUT=${3:-$HOME/d2d-lab/pr/necskill-out/m$SEED}
mkdir -p "$OUT"
. "$HERE/nec_env.sh"
export PYTHONPATH=$HERE:$D2/tools/emu
cd "$D2/tools/emu"
$PY "$HERE/nec_missile_hit.py" gen "$N" "$SEED" > "$OUT/cases.tsv"
nice $PY "$HERE/nec_missile_hit.py" game "$OUT/cases.tsv" > "$OUT/game.tsv" 2> "$OUT/game.err"
cat "$OUT/game.err" | tail -2
nice "$HOME/d2d-lab/pr/necskill-out/nec_missile_test" --missile-hit "$OUT/cases.tsv" 2> "$OUT/d2d.err" | grep -E '^[0-9]+ ' > "$OUT/d2d.tsv"
wc -l "$OUT/cases.tsv" "$OUT/game.tsv" "$OUT/d2d.tsv" | head -3
echo "DIFF $(diff "$OUT/game.tsv" "$OUT/d2d.tsv" | grep -c '^<')"
