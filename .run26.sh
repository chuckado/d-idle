#!/bin/sh
# usage: run26.sh BINNAME LABEL
cd ~/labrun/herdr/code/d2lod-wt/drills26 || exit 1
go build -o ~/d2d-lab/drills26/d2drill ./cmd/d2drill || exit 1
F=$HOME/d2d-lab/fork/bin
D2D_AGENT=drills26 D2D_BIN=$F/$1 timeout 300 tools/agentrun.sh drills26 --ttl 330 -- tools/d2dslot.sh run -n 0 -t 300 -- $HOME/d2d-lab/drills26/d2drill -group U -repeat 1 -bin $F/$1 -out /tmp/d26out -label $2 -v > /tmp/$2.log 2>&1
grep -E "^DRILL|^END" /tmp/$2.log | cut -c1-3000
