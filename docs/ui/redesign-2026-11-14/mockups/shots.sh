#!/bin/sh
# Pictures of the static mockup (design phase). Runs ON the worker host through labrun, from the repo root:
#   labrun <worker> --pull docs/ui/redesign-2026-11-14/mockups/shots -- sh docs/ui/redesign-2026-11-14/mockups/shots.sh
# The mockup file has no doctype of its own (the Artifact page adds one), so a wrapped copy is rendered.
set -u
d=docs/ui/redesign-2026-11-14/mockups
chrome=$(command -v google-chrome-stable || command -v chromium || command -v chromium-browser) || { echo "no Chrome" >&2; exit 2; }
tmp=$(mktemp -d); trap 'rm -rf "${tmp:?}"' EXIT
for m in watch-orders character-screen; do
{ printf '<!doctype html><html data-theme="dark"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body>'; cat "$d/$m.html"; printf '</body></html>'; } >"$tmp/$m.html"
done
mkdir -p "$d/shots"
for s in "watch-orders mockup 1440 1100" "watch-orders mockup 900 1500" "watch-orders mockup 390 1700" "character-screen character 1440 900" "character-screen character 900 1300" "character-screen character 390 1500"; do
	set -- $s
	timeout 40 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$tmp/p$2$3" \
		--force-device-scale-factor=1 --window-size="$3,$4" --virtual-time-budget=1500 --screenshot="$d/shots/$2-$3.png" "file://$tmp/$1.html" >/dev/null 2>&1
done
ls -la "$d/shots"
