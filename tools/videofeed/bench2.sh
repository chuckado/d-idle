#!/bin/sh
# On the builder: putImageData against a WebGL texture upload in headless Chromium. Writes /tmp/vf3/bench-result.txt.
set -u
mkdir -p /tmp/vf3; rm -f /tmp/vf3/bench-result.txt
node tools/videofeed/serve.js /tmp/vf3 18932 &
srv=$!
sleep 1
timeout 100 /usr/bin/chromium --headless=new --no-sandbox --user-data-dir=/tmp/vf3/profile --use-gl=angle --use-angle=swiftshader --enable-unsafe-swiftshader http://127.0.0.1:18932/bench2.html >/dev/null 2>&1 &
cpid=$!
i=0
while [ $i -lt 90 ] && [ ! -f /tmp/vf3/bench-result.txt ]; do sleep 1; i=$((i + 1)); done
kill $cpid 2>/dev/null; kill $srv 2>/dev/null
cat /tmp/vf3/bench-result.txt 2>/dev/null || echo "no result in 90 s"
rm -rf /tmp/vf3/profile
