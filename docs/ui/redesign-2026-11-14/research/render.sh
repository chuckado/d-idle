#!/bin/sh
# Renders the v2 mockups at the exact acceptance viewports (design phase, not a test). Runs ON the builder through
# labrun, from the repo root, never on the worker host that runs the games:
#   labrun <builder> --pull docs/ui/redesign-2026-11-14/research/iterations -- sh docs/ui/redesign-2026-11-14/research/render.sh <tag> [<tag>...]
# Sources: research/iterations/<tag>/src/*.html when present (a frozen iteration), else mockups/v2/*.html, which are
# then copied there so every iteration keeps the page it was rendered from. Headless Chromium on a private Xvfb
# display, as tools/uishots.sh does. The mockup files have no doctype (the Artifact page adds one): a wrapped copy
# is rendered.
set -u
d=docs/ui/redesign-2026-11-14
chrome=${CHROME:-$(command -v chromium || command -v chromium-browser || command -v google-chrome-stable)} || { echo "no Chromium" >&2; exit 2; }
tmp=$(mktemp -d); xp=
trap 'kill $xp 2>/dev/null; rm -rf "${tmp:?}"' EXIT INT TERM
if command -v Xvfb >/dev/null; then timeout 400 Xvfb :247 -screen 0 1920x1200x24 -nolisten tcp >/dev/null 2>&1 & xp=$!; DISPLAY=:247; export DISPLAY; sleep 1; fi
for tag in "$@"; do
out=$d/research/iterations/$tag
mkdir -p "$out/src"
for m in watch character; do
	[ -f "$out/src/$m.html" ] || cp "$d/mockups/v2/$m.html" "$out/src/$m.html" || continue
	{ printf '%s' '<!doctype html><html data-theme="dark"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><style>html{color-scheme:dark}body{margin:0;font:14px system-ui}[hidden]{display:none!important}</style></head><body>'
	  cat "$out/src/$m.html"
	  printf '%s' '</body></html>'; } >"$tmp/$m.html"
	for s in "1440 900 " "1920 1080 " "1366 768 " "390 844 " "1440 900 #tip"; do
		set -- $s
		name=$m-$1x$2; [ -n "${3:-}" ] && name=$name-tip
		[ -n "${3:-}" ] && [ "$m" = watch ] && continue
		timeout 45 nice -n 19 "$chrome" ${HEADLESS:---headless=new} --no-sandbox --disable-gpu --disable-dev-shm-usage --hide-scrollbars --no-first-run --user-data-dir="$tmp/p-$name" \
			--force-device-scale-factor=1 --window-size="$1,$2" --virtual-time-budget=1500 --screenshot="$out/$name.png" "file://$tmp/$m.html${3:-}" >/dev/null 2>&1
		echo "$name rc=$?"
	done
done
ls "$out"
done
