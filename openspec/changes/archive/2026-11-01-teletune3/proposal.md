## Why

DRILLS25 left three loose ends of card teletune2: the border leg costs 9 ticks with a Teleport key bound (717 against 708, 0 hops), the stall escape (`stalledHere`) was never exercised live, and one process running D22-PC then D22-P gave 0 hops because the navigator keeps its stall memory. docs/teletune3.md has the analysis.

## What Changes

- internal/nav: `Navigator.ResetStallMemory()` (stall spots, hard blocks, stalled gates, declined exits, refused hop cells, search memory; the bot never calls it); a `why` field on the `nav plan` debug line (first, goal, map, monsters, deviation, partial). No bot behaviour change.
- cmd/d2drill, internal/drill: `D22()` resets the memory before every drill; drill D22-S (stall escape on one level with a strict leg rule) and D22-BL (border leg three times per arm with tick-stamped plan lines). Not run.
- Tests: the stall rule through the real `noteStuck`, the reset, and that a bound key does not change a border leg.

## Impact

Offline only. Awaits the DRILLS26 recipe in docs/teletune3.md. Build-flexibility: nothing here names a skill id or class; the tests run with two keys; the drills resolve the key through `Fighter.TeleportKey()`.

Overlap (declared): card m2-roam (deferred) owns internal/nav/exit.go and roam.go and card m1-liveproof owns cmd/d2drill/** and internal/drill/**; this card edits internal/nav/walk.go (log field), internal/nav/stuckmem.go, cmd/d2drill/d22_lab.go and internal/drill/d22.go without listing them in owns.
