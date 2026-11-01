## Why

The skill traces (sk-teleport, sk-mana, sk-missile, sk-armor in docs/research) list drill recipes that had never been run live. This card runs the ones the lab verbs of FORKDBG allow on a hidden d2d slot and compares d2d with the trace's game.exe expectation.

## What Changes

- New drill group `R` in cmd/d2drill and internal/drill (new files d15.go, d15_lab.go, d15b_lab.go; a small hook in cmd/d2drill/main.go for the group switch and the end-of-run goto town): D15-T1..T5, T7, M2, M3, M4, M6, K1, K2, E1.
- Result log `docs/probes/drills/2026-11-01-drills15-full.txt` and a DRILLS15 section in docs/probes/drills/README.md; rows in docs/primitives.md.
- No product code changes (deviations are reported, not fixed).
