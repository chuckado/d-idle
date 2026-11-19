#!/bin/bash
# CAIN-L1 (stone sequence), ON THE WORKER from the synced worktree. usage: run.sh LABEL 'verb;verb;...' [SLOT]
# One lab-verbs d2d (lockstep, no save) on a d2d slot, throwaway Sorceress; d2drill -probe sends the verbs (debug verbs
# are ledgered with a reason). Output goes to /tmp/cainl1-out/LABEL.txt.
export PATH=$PATH:$HOME/.local/bin
B=$HOME/d2d-lab/fork/bin/d2d.lab3-labverbs-f75816c6-linuxbuild
go build -o /tmp/cainl1-d2drill ./cmd/d2drill || exit 1
mkdir -p /tmp/cainl1-out
D2D_AGENT=cainl1 D2D_BIN=$B exec tools/agentrun.sh cainl1 --ttl 400 -- tools/d2dslot.sh run -n "${3:-3}" -t 360 -- \
  nice -n 19 /tmp/cainl1-d2drill -seed "${SEED:-11}" -probe "$2" -label "$1" -out /tmp/cainl1-out -bin $B
