#!/bin/sh
# Vault page pictures from the fixture viewer (card vaultclean): tools/uishots/vaultshots.sh PREFIX [state...] (heavy: headless Chrome, long lane).
cd "$(dirname "$0")/../.." || exit 2
. tools/lib/heavy-guard.sh
heavy_guard vaultshots || exit $?
bin=$(mktemp -d); trap 'kill $sp 2>/dev/null; rm -rf "$bin"' EXIT
chrome=$(command -v google-chrome-stable || command -v chromium) || exit 2
go build -o "$bin/uishots" ./tools/uishots || exit 2
timeout 200 "$bin/uishots" serve -addr 127.0.0.1:18977 >"$bin/s.log" 2>&1 &
sp=$!; sleep 3
out=${OUT:-docs/ui/vault-cleanup}; mkdir -p "$out"
for sz in ${SIZES:-1366x768 1920x1080 390x844}; do
 for st in ${STATES:-revealing}; do
  timeout 60 $chrome --headless=new --no-sandbox --disable-gpu --hide-scrollbars --user-data-dir=$bin/p$sz --window-size=${sz%x*},${sz#*x} --virtual-time-budget=6000 --screenshot=$out/$1-$st-$sz.png "http://127.0.0.1:18977/watch/vault?vault=$st&theme=dark" >/dev/null 2>&1
 done
done
ls "$out"
