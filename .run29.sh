#!/bin/sh
# usage: .run29.sh BINNAME LABEL GROUP TTL [only]   (from inside the synced worktree on the worker host)
cd "$(dirname "$0")" || exit 1
mkdir -p ~/d2d-lab/drills29
go build -o ~/d2d-lab/drills29/d2drill ./cmd/d2drill || exit 1
F=$HOME/d2d-lab/fork/bin
O=""; [ -n "$5" ] && O="-only $5"
D29_MON=$D29_MON D2D_AGENT=drills29 D2D_BIN=$F/$1 timeout $4 tools/agentrun.sh drills29 --ttl $4 -- tools/d2dslot.sh run -n ${SLOT:-0} -t $4 -- $HOME/d2d-lab/drills29/d2drill -group $3 -repeat 1 -bin $F/$1 -out /tmp/d29out -label $2 -v $O > /tmp/d29-$2.log 2>&1
grep -E "^DRILL|^END" /tmp/d29-$2.log | cut -c1-6000
