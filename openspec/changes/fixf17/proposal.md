## Why

DRILLS9 finding F17: `DW.m1` in the full F group ends BLOCKED in run 1 with `nav: timeout: walk did not finish in 2m0s` (5 of 5 invocations) and passes in run 2. Reading the DRILLS9 logs, the cause is in the harness, not the product: the lab stepped the game with `Client.Step` directly, which the Pacer (the navigator's clock under lockstep) does not see, so the Pacer's virtual time lagged the game and the next walk's first Sleep caught up all lagging ticks at once (docs/fixf17.md).

## What Changes

- cmd/d2drill/live.go: `lab.Tick` re-syncs the Pacer after each direct step (`syncPacer`: one tick through the Pacer, returns the lag it absorbed); the object drills' `Client.Step(250)` go through `Tick`.
- cmd/d2drill/wp_lab.go: the waypoint trip syncs the Pacer first and logs `drill: the Pacer lagged the game before a waypoint trip` with the lag (diagnostic for the live proof).
- cmd/d2drill/pacer_sync_test.go: regression (fails before, passes after) on the fake lockstep d2d.
- No product code and no numeric gate changed (no delta spec).

Overlap (declared, not owned): cmd/d2drill/** belongs to the active card m1-liveproof (docs/OWNERS). This card edits live.go, wp_lab.go, field_lab.go (one line) and adds pacer_sync_test.go there only for the Pacer sync; to avoid an ownership overlap it owns only docs/fixf17.md.
