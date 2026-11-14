#!/bin/sh
# Renders the v2 mockups at the exact acceptance viewports (design phase, not a test). Runs ON the builder through
# labrun, from the repo root, never on the worker host that runs the games:
#   labrun <builder> --pull docs/ui/redesign-2026-11-14/research/iterations -- sh docs/ui/redesign-2026-11-14/research/render.sh <tag>
# Headless Chromium, no display needed. The mockup files have no doctype (the Artifact page adds one), so a wrapped
# copy is rendered. After each render the script prints, per viewport, whether the page scrolls (it must not).
set -u
tag=${1:-it}
d=docs/ui/redesign-2026-11-14
out=$d/research/iterations/$tag
chrome=$(command -v chromium || command -v chromium-browser || command -v google-chrome-stable) || { echo "no Chromium" >&2; exit 2; }
tmp=$(mktemp -d); trap 'rm -rf "${tmp:?}"' EXIT
mkdir -p "$out"
for m in watch character; do
	[ -f "$d/mockups/v2/$m.html" ] || continue
	{ printf '<!doctype html><html data-theme="dark"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><style>html{color-scheme:dark}body{margin:0;font:14px system-ui}img{max-width:100%}[hidden]{display:none!important}</style></head><body>'
	  cat "$d/mockups/v2/$m.html"
	  printf '<script>addEventListener("load",function(){var e=document.scrollingElement;document.title="SCROLL "+(e.scrollHeight-innerHeight)+" "+(e.scrollWidth-innerWidth)})</script></body></html>'; } >"$tmp/$m.html"
	for s in "1440 900" "1920 1080" "1366 768" "390 844"; do
		set -- $s
		timeout 60 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --hide-scrollbars --no-first-run --user-data-dir="$tmp/p-$m-$1" \
			--force-device-scale-factor=1 --window-size="$1,$2" --virtual-time-budget=1500 --screenshot="$out/$m-$1x$2.png" "file://$tmp/$m.html" >/dev/null 2>&1
		t=$(timeout 60 nice -n 19 "$chrome" --headless=new --no-sandbox --disable-gpu --no-first-run --user-data-dir="$tmp/p-$m-$1" \
			--window-size="$1,$2" --virtual-time-budget=1500 --dump-dom "file://$tmp/$m.html" 2>/dev/null | grep -o '<title>SCROLL[^<]*' | head -1)
		echo "$m $1x$2: ${t#<title>} (overflow px down, right; 0 0 = fits)"
	done
done
ls "$out"
