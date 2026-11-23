#!/bin/sh
# Counts the long-lived requests the /watch page holds (docs: the viewer-connections card). Runs on the worker host:
#   tools/connprobe/run.sh [sse|nosse]    serves the uishots fixture viewer, loads /watch in headless Chrome, reloads twice
set -u
cd "$(dirname "$0")/../.." || exit 2
bin=$(mktemp -d); trap 'kill $sp 2>/dev/null; rm -rf "$bin"' EXIT
chrome=$(command -v google-chrome-stable || command -v chromium) || { echo "connprobe: no Chrome" >&2; exit 2; }
go build -o "$bin/uishots" ./tools/uishots || exit 2
UISHOTS_SSE=1 timeout 300 "$bin/uishots" serve -addr 127.0.0.1:18952 >"$bin/serve.log" 2>&1 &
sp=$!; sleep 3
for u in '/watch?stream=live&pal=1' '/watch?stream=live'; do
	echo "== $u"; CHROME=$chrome TMPDIR=$bin timeout 120 node --experimental-websocket tools/connprobe/conns.mjs "http://127.0.0.1:18952$u" 2 12000 1
done
echo "== two tabs"; CHROME=$chrome TMPDIR=$bin timeout 120 node --experimental-websocket tools/connprobe/conns.mjs "http://127.0.0.1:18952/watch?stream=live&pal=1" 0 12000 2
