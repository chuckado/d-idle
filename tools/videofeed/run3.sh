#!/bin/sh
# On the builder, from the repo root, with frames recorded by run2.sh in /tmp/vf2-frames: write a 400-frame wire stream and
# run the pacing bench (bench3.html) in headless Chromium: the old present-on-arrival against the presentation clock.
# Writes /tmp/vf4/bench3-result.txt. Frames stay in /tmp, nothing is committed.
set -u
mkdir -p /tmp/vf4; rm -f /tmp/vf4/bench3-result.txt
[ -f /tmp/vf2-frames/scenes.txt ] || python3 tools/videofeed/record2.py /tmp/vf2-frames || exit 1
D2FEED_FRAMES=/tmp/vf2-frames D2FEED_MSGS=/tmp/vf4/stream.pf D2FEED_MAX=400 go test -count=1 -run TestWriteWire ./internal/watch/palfeed || exit 1
node tools/videofeed/stream3.js /tmp/vf4 18933 &
srv=$!
sleep 1
timeout 120 /usr/bin/chromium --headless=new --no-sandbox --disable-gpu --user-data-dir=/tmp/vf4/profile http://127.0.0.1:18933/bench3.html >/dev/null 2>&1 &
cpid=$!
i=0
while [ $i -lt 100 ] && [ ! -f /tmp/vf4/bench3-result.txt ]; do sleep 1; i=$((i + 1)); done
kill $cpid 2>/dev/null; kill $srv 2>/dev/null
cat /tmp/vf4/bench3-result.txt 2>/dev/null || echo "no result in 100 s"
rm -rf /tmp/vf4/profile
