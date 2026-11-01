#!/bin/sh
cd "$(dirname "$0")" || exit 1
go build -o ~/d2d-lab/drills29/d2drill ./cmd/d2drill || exit 1
F=$HOME/d2d-lab/fork/bin
D2D_AGENT=drills29 D2D_BIN=$F/d2d.lab-armorhit timeout 100 tools/agentrun.sh drills29 --ttl 100 -- tools/d2dslot.sh run -n 0 -t 100 -- $HOME/d2d-lab/drills29/d2drill -group W -repeat 1 -bin $F/d2d.lab-armorhit -out /tmp/d29out -probe "$1" 2>&1 | cut -c1-1500
