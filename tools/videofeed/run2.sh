#!/bin/sh
# On the builder, from the repo root: record the walking-heavy sequence (record2.py) and run the bandwidth comparison
# (TestCompare). Writes /tmp/vf2-report.txt; frames stay in /tmp/vf2-frames (never committed).
#   tools/videofeed/run2.sh [D2D_BIN]
set -u
d2d=${1:-/data/build/d2d.lab3-labverbs-f75816c6-d2d/apps/d2d/d2d}
if [ "${VF_REUSE:-0}" != 1 ] || [ ! -f /tmp/vf2-frames/scenes.txt ]; then
	rm -rf /tmp/vf2-frames
	python3 tools/videofeed/record2.py /tmp/vf2-frames "$d2d" || exit 1
fi
D2FEED_FRAMES=/tmp/vf2-frames D2FEED_OUT=/tmp/vf2-report.txt go test -count=1 -timeout 30m -run TestCompare -v ./internal/watch/palfeed >/tmp/vf2-go.log 2>&1
tail -3 /tmp/vf2-go.log
cat /tmp/vf2-report.txt
