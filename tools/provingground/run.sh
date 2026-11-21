#!/bin/bash
# Proving Ground runner wrapper, ON THE BUILDER (long lane), from the synced worktree (docs/process/proving-ground.md).
#   labrun BUILDER --lane long --detach --notify --pull provingground-out -- tools/provingground/run.sh [d2scenario run flags]
# Builds d2bot and d2scenario, checks memory, then runs the selected scenarios in lab slots (never the worker host, never the head box).
#   run.sh -tier fast                       the merge-gate subset
#   run.sh -id cain.scroll -class Paladin -seed 11 -parallel 1     the one-scenario dev loop
#   run.sh -dry-run -tier full              what would run
#   run.sh -journal ...                     also write the d2dctl journal of each bot run (cmd/d2ddet)
#   run.sh determinism-run -n 3 -id town.resupply -class Paladin -seed 11 -parallel 1    N repetitions, journals on, compared with d2ddet
# Env: PG_D2D_BIN (lab d2d, default the survival lab build), PG_DATA (MPQ dir), PG_MIN_MB (memory guard, default 2500).
set -u
here=$(cd "$(dirname "$0")" && pwd); repo=$(cd "$here/../.." && pwd); cd "$repo" || exit 2
export PATH=$PATH:/opt/go/bin:$HOME/.local/bin
case " $* " in *" -dry-run "*) dry=1 ;; *) dry= ;; esac
avail=$(awk '/^MemAvailable:/{print int($2/1024)}' /proc/meminfo)
if [ -z "$dry" ] && [ "${avail:-0}" -lt "${PG_MIN_MB:-2500}" ]; then echo "provingground: only ${avail} MB available, need ${PG_MIN_MB:-2500}" >&2; exit 75; fi
W=${PG_DIR:-/data/d2d-lab/provingground}; mkdir -p "$W/bin"
go build -o "$W/bin/d2scenario" ./cmd/d2scenario || exit 1
go build -o "$W/bin/d2ddet" ./cmd/d2ddet || exit 1 # determinism-run compares journals with it
[ -n "$dry" ] || go build -o "$W/bin/d2bot" ./cmd/d2bot || exit 1
export D2D_BIN=${PG_D2D_BIN:-/data/d2d-lab/survhold/bin/d2d.lab} D2D_DATA=${PG_DATA:-/data/d2d-lab/data} D2BOT_BIN=$W/bin/d2bot
out=provingground-out/$(date +%Y%m%d-%H%M%S)
sub=run; if [ "${1:-}" = determinism-run ]; then sub=determinism-run; shift; fi # N repetitions, journals compared
exec "$W/bin/d2scenario" $sub -slot "$repo/tools/d2dslot.sh" -out "$out" "$@"
