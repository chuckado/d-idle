#!/bin/sh
# usage: run28.sh BINNAME LABEL GROUP TTL  (from inside the synced worktree on the worker)
cd ~/labrun/herdr/code/d2lod-wt/drills28 || exit 1
mkdir -p ~/d2d-lab/drills28
go build -o ~/d2d-lab/drills28/d2drill ./cmd/d2drill || exit 1
F=$HOME/d2d-lab/fork/bin
D2D_AGENT=drills28 D2D_BIN=$F/$1 timeout $4 tools/agentrun.sh drills28 --ttl $4 -- tools/d2dslot.sh run -n 0 -t $4 -- $HOME/d2d-lab/drills28/d2drill -group $3 -repeat 1 -bin $F/$1 -out /tmp/d28out -label $2 -v > /tmp/$2.log 2>&1
grep -E "^DRILL|^END" /tmp/$2.log | cut -c1-4000
