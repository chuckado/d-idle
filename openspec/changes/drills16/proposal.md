## Why

Two cards were open awaiting a live proof: fixmouth (arrival step-away radius, docs/fixmouth.md) and fixbuff (armor recast inside a long fight, d2d buff states, docs/fixbuff.md).

## What Changes

Live drills only: new group R (internal/drill/d16.go, cmd/d2drill/d16_lab.go) for the fixbuff items, the DRILLS14 DW.m3 / DW.m4 drills rerun for fixmouth, a DRILLS16 section in docs/probes/drills/README.md. No product code.

## Impact

Overlap (declared): cmd/d2drill/main.go gets a `case "R"` and the end-of-run goto-town condition names "R" (two lines, as DRILLS13 did for "Q").
