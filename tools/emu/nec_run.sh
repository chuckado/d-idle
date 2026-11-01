#!/bin/bash
# necskill: calc differential on the worker host. Usage (from the worktree root):
#   tools/emu/nec_run.sh N SEED OUTDIR
set -e
HERE=$(cd "$(dirname "$0")" && pwd)
N=${1:-3000}; SEED=${2:-1}; OUT=${3:-$HOME/d2d-lab/pr/necskill-out/run}
mkdir -p "$OUT"
. "$HERE/nec_env.sh"
export PYTHONPATH=$HERE:$D2/tools/emu
cd "$D2/tools/emu"
$PY "$HERE/nec_calc.py" gen "$N" "$SEED" > "$OUT/cases.tsv"
nice $PY "$HERE/nec_calc.py" game "$OUT/cases.tsv" > "$OUT/game.tsv" 2> "$OUT/game.err"
nice "$HOME/d2d-lab/pr/necskill-out/nec_test" --calc "$OUT/cases.tsv" 2> "$OUT/d2d.err" | grep -E "^[0-9]+ [-0-9x]" > "$OUT/d2d.tsv"
wc -l "$OUT/cases.tsv" "$OUT/game.tsv" "$OUT/d2d.tsv"
if diff -q "$OUT/game.tsv" "$OUT/d2d.tsv" >/dev/null; then echo "DIFF 0"; else echo "DIFF $(diff "$OUT/game.tsv" "$OUT/d2d.tsv" | grep -c '^<')"; fi
