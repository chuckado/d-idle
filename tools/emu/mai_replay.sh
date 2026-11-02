#!/bin/bash
# monsterai: replay the dump (tests/mai_cases.inc, written by mai_run.sh) through d2d's rules::mon_think; prints `mai AI cases differ` per AI.
SRC=$HOME/d2d-lab/pr/monsterai-src; OUT=$HOME/d2d-lab/pr/monsterai-out
HERE=$(cd "$(dirname "$0")" && pwd)
cd "$SRC" || exit 1
git checkout tests/test_monsters.cpp 2>/dev/null
python3 "$HERE/mai_patch.py" . || exit 1
"$HERE/mai_build.sh" "$SRC" >/dev/null || { tail -20 "$OUT/build.log"; exit 1; }
"$OUT/build/tests/test_monsters" 2> "$OUT/replay-stderr.txt"; rc=$?
echo "rc=$rc stderr lines: $(wc -l < "$OUT/replay-stderr.txt")"
git checkout tests/test_monsters.cpp
