#!/bin/bash
# monsterai: the two runs of docs/research/monster-ai-trace.md, worker host. Run A: seed 20261101, commands as the scene draws them
# (QuillRat / Fetish none); run B: seed 7, MAI_COMMAND=none. 1500 cases per AI, 25 AIs. About 6 minutes (the test compiles twice).
HERE=$(cd "$(dirname "$0")" && pwd); OUT=$HOME/d2d-lab/pr/monsterai-out; SRC=$HOME/d2d-lab/pr/monsterai-src
for run in "A 20261101 x" "B 7 none"; do
  set -- $run
  MAI_COMMAND=$([ "$3" = none ] && echo none) TAILN=0 "$HERE/mai_run.sh" all 1500 "$2" > /dev/null 2>&1
  "$HERE/mai_replay.sh" > /dev/null 2>&1
  (cd "$SRC" && "$OUT/build/tests/test_monsters" 2> "$OUT/replay-$1-stderr.txt" | grep '^mai' > "$OUT/replay-$1.txt")
  echo "run $1: $(grep -c . "$OUT/replay-$1.txt") lines, $(grep TOTAL "$OUT/replay-$1.txt")"
done
