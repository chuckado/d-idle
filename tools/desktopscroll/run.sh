#!/bin/sh
# Desktop scroll check (card vaultdesktop): the Vault page is one scrolling document at every desktop size.
# Heavy: headless Chrome. On the builder run it in the long lane (labrun builder --lane long -- tools/desktopscroll/run.sh); the guard refuses otherwise.
#   tools/desktopscroll/run.sh [old]   with "old" the 100dvh frame is put back and the check must FAIL (exit 0 if it does).
set -u
cd "$(dirname "$0")/../.." || exit 2
. tools/lib/heavy-guard.sh
heavy_guard desktopscroll || exit $?
bin=$(mktemp -d); trap 'kill $sp 2>/dev/null; rm -rf "$bin"' EXIT
chrome=$(command -v google-chrome-stable || command -v chromium) || { echo "desktopscroll: no Chrome" >&2; exit 2; }
go build -o "$bin/uishots" ./tools/uishots || exit 2
port=${DESKTOPSCROLL_PORT:-18971}
timeout 400 "$bin/uishots" serve -addr 127.0.0.1:$port >"$bin/serve.log" 2>&1 &
sp=$!; sleep 3
d=${1:-}
OLD=${d:+1} CHROME=$chrome TMPDIR=$bin timeout 360 node --experimental-websocket tools/desktopscroll/vault.mjs "http://127.0.0.1:$port" "${DESKTOPSCROLL_SIZES:-1366x768 1920x1080 1280x720 900x700}"
rc=$?
if [ -n "$d" ]; then [ $rc = 1 ] && { echo "the old CSS is correctly caught"; exit 0; }; echo "the old CSS was NOT caught" >&2; exit 1; fi
exit $rc
