#!/bin/sh
# On the builder, from the repo root: record real d2d frames, replay them through the codec, bench node and a headless
# browser. Writes /tmp/vf-report.txt (and leaves /tmp/vf-frames, never committed).
#   tools/videofeed/run.sh [D2D_BIN]
set -u
d2d=${1:-/data/build/d2d.lab3-labverbs-f75816c6-d2d/apps/d2d/d2d}
# VF_REUSE=1: keep the frames recorded by an earlier run
rm -f /tmp/vf-report.txt /tmp/vf-frames/bench-result.txt
if [ "${VF_REUSE:-0}" != 1 ] || [ ! -f /tmp/vf-frames/scenes.txt ]; then
	rm -rf /tmp/vf-frames
	python3 tools/videofeed/record.py /tmp/vf-frames "$d2d" || exit 1
fi
D2FEED_FRAMES=/tmp/vf-frames D2FEED_OUT=/tmp/vf-report.txt D2FEED_MSGS=/tmp/vf-frames/stream.pf go test -count=1 -run TestReplay -v ./internal/watch/palfeed >/tmp/vf-go.log 2>&1
tail -3 /tmp/vf-go.log
{ echo; node tools/videofeed/replay.js /tmp/vf-frames/stream.pf; } >>/tmp/vf-report.txt 2>&1
node tools/videofeed/serve.js /tmp/vf-frames 18931 &
srv=$!
sleep 1
if [ -x /usr/bin/chromium ]; then
	timeout 100 /usr/bin/chromium --headless=new --no-sandbox --disable-gpu --user-data-dir=/tmp/vf-chromium http://127.0.0.1:18931/bench.html >/dev/null 2>&1 &
	cpid=$!
	i=0
	while [ $i -lt 90 ] && [ ! -f /tmp/vf-frames/bench-result.txt ]; do sleep 1; i=$((i + 1)); done
	kill $cpid 2>/dev/null
	{ echo; cat /tmp/vf-frames/bench-result.txt 2>/dev/null || echo "browser bench: no result in 90 s"; } >>/tmp/vf-report.txt
	rm -rf /tmp/vf-chromium
fi
kill $srv 2>/dev/null
echo "report: /tmp/vf-report.txt"
