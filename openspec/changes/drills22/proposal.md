## Why

Card fixtelekey wired Teleport as a capability of the character (Fighter bind, level rule) but was proven offline only. DRILLS22 is the live proof on a hidden d2d slot before the merge.

## What Changes

- New drill group T (internal/drill/d22.go, cmd/d2drill/d22_lab.go): Act 2 legs with and without Teleport, the bind readback, town and level rules, refused-hop bound.
- Results and findings in docs/probes/drills/README.md (DRILLS22) and the logs `2026-11-01-drills22-*.txt`.

## Impact

Harness and docs only; no bot code changes. Overlap (declared): this card adds internal/drill/d22.go, cmd/d2drill/d22_lab.go and the logs docs/probes/drills/2026-11-01-drills22-*.txt without listing them in owns (standing defaults of the drills agent and m1-liveproof), and edits the group switch and the end-of-run goto town in cmd/d2drill/main.go.
