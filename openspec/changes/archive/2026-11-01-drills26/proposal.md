## Why

no observable behaviour change: harness and docs only (live drill, no bot code).

The Frozen Armor recast (fixrecast, fixbuff) was proven against a harmless held monster. DRILLS26 runs the bot's own combat loop against a real, replaced zombie pack for 300 game seconds on the fork binary that follows game.exe's freeze rule (lab-armorhit) and on the previous one (lab-nexthit2).

## What Changes

- d2drill: group U, drill D26-A (internal/drill/d26.go, cmd/d2drill/d26_lab.go), the binary named in the result.
- Results and findings in docs/probes/drills/README.md (DRILLS26) and `2026-11-01-drills26-*.txt`.

## Impact

Harness and docs only. Overlap (declared): cmd/d2drill/main.go (group switch, binName).
