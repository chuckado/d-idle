#!/bin/sh
# Phone scroll check (card mobilequests): on a 412x915 touch phone every cockpit tab scrolls with the page, no nested scroll trap.
# Runs on the worker host (headless Chrome): tools/mobilescroll/run.sh [defect]    with a defect name the check must FAIL (exit 0 if it does).
set -u
cd "$(dirname "$0")/../.." || exit 2
bin=$(mktemp -d); trap 'kill $sp 2>/dev/null; rm -rf "$bin"' EXIT
chrome=$(command -v google-chrome-stable || command -v chromium) || { echo "mobilescroll: no Chrome" >&2; exit 2; }
go build -o "$bin/uishots" ./tools/uishots || exit 2
port=${MOBILESCROLL_PORT:-18961}
timeout 300 "$bin/uishots" serve -addr 127.0.0.1:$port >"$bin/serve.log" 2>&1 &
sp=$!; sleep 3
d=${1:-}
url="http://127.0.0.1:$port/watch?first=0&needs=0&quests=running"
[ -n "$d" ] && url="$url&defect=$d"
CHROME=$chrome TMPDIR=$bin timeout 240 node --experimental-websocket tools/mobilescroll/probe.mjs "$url" ${MOBILESCROLL_SIZE:-412x915}
rc=$?
if [ -n "$d" ]; then [ $rc = 1 ] && { echo "defect $d correctly fails"; exit 0; }; echo "defect $d was NOT caught" >&2; exit 1; fi
exit $rc
