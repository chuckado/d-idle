## Why

no observable behaviour change: harness and docs only (live proof drills, no bot code).

Card teletune changed when the navigator hops with Teleport but was proven offline only. DRILLS24 is the live proof on a hidden d2d slot before the merge.

## What Changes

- d2drill: flags `-tele-min-leg` and `-tele-reserve`, `TeleportCost` wiring in d22wire, per-hop tick log, pocket drills D22-PC and D22-P in group T.
- Results and findings in docs/probes/drills/README.md (DRILLS24) and the logs `2026-11-01-drills24-*.txt`.

## Impact

Harness and docs only. Overlap (declared): cmd/d2drill/main.go (flags and the recorder tick source), cmd/d2drill/rec.go, cmd/d2drill/d22_lab.go and internal/drill/d22.go, files of card drills22 that this branch carries.
