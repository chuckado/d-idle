#!/bin/sh
# Phone scroll check (card mobilequests): on a 412x915 touch phone every cockpit tab scrolls with the page, no nested scroll trap.
# Runs on the worker host (headless Chrome): tools/mobilescroll/run.sh [old]    with "old" the pre-fix CSS is put back and the check must FAIL (exit 0 if it does).
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
OLD=${d:+1} CHROME=$chrome TMPDIR=$bin timeout 240 node --experimental-websocket tools/mobilescroll/probe.mjs "$url" ${MOBILESCROLL_SIZE:-412x915}
rc=$?
vurl="http://127.0.0.1:$port/watch/vault?vault=pop"
OLD=${d:+1} CHROME=$chrome TMPDIR=$bin timeout 240 node --experimental-websocket tools/mobilescroll/vault.mjs "$vurl" ${MOBILESCROLL_SIZE:-412x915}
rc2=$?
[ $rc = 0 ] && rc=$rc2; [ $rc2 = 1 ] && rc=1
if [ -n "$d" ]; then [ $rc = 1 ] && { echo "the old CSS is correctly caught"; exit 0; }; echo "the old CSS was NOT caught" >&2; exit 1; fi
exit $rc
