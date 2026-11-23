#!/bin/sh
# Before/after pictures at 412x915 into docs/ui (fixture viewer, no hero data): tools/mobilescroll/shots.sh OUTDIR [old]
set -u
cd "$(dirname "$0")/../.." || exit 2
out=${1:?outdir}; tag=after; [ "${2:-}" = old ] && tag=before
bin=$(mktemp -d); trap 'kill $sp 2>/dev/null; rm -rf "$bin"' EXIT
go build -o "$bin/uishots" ./tools/uishots || exit 2
timeout 200 "$bin/uishots" serve -addr 127.0.0.1:18963 >"$bin/serve.log" 2>&1 &
sp=$!; sleep 3
OLD=${2:+1} TMPDIR=$bin timeout 150 node --experimental-websocket tools/mobilescroll/shots.mjs http://127.0.0.1:18963 "$out" $tag
