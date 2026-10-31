## Why

fixstall3 (adaptive approach clearance, walking portal entry, F12 diagnostics) was proven offline only. DRILLS8 runs H.pause, D6.d and D-sync live against it and does a regression pass of the whole drill suite.

## What Changes

- Results in docs/probes/drills (drills8 files, README section DRILLS8), primitives rows, a `live:` line on the fixstall3 card.
- Harness edits only in cmd/d2drill and internal/drill (listed in the README section). No product code changes: failures are findings.
- docs/OWNERS: the drills8 card owns only its own change directory; the drill paths stay with m1-liveproof (no new overlap).
