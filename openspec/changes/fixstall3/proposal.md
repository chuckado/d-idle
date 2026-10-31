## Why

DRILLS7 (docs/probes/drills/README.md, F10 to F12) showed the fixstall2 changes did not work live: H.pause stalls 3 of 6 tries (the grid route finds no path with clearance 1 from the stall cells), D6.d still pre-walks to the portal (the walking input is never wired), D-sync.wp ignores zombies after 12 approaches in one of two runs.

## What Changes

- internal/d2dctl/approach_path.go: adaptive clearance (1 first, 0 as fallback), regression test on the real camp grid fixture.
- cmd/d2bot/d2dwire.go, cmd/d2drill/live.go: the town input is `d2dtown.Input` (WalkingInteractor, BeltPut, MoveGold), test in internal/d2dtown.
- internal/combat/fight.go: the approaches-ignore reason carries the geometry (F12 cause is UNVERIFIED, next probe in docs/fixstall3.md).
- No numeric gate changes (no delta).

Overlap (declared, not edited): internal/d2dctl/approach_path.go, approach_path_test.go and internal/d2dtown/adapter.go (unchanged here) are owned by the open card fixstall2; cmd/d2bot/d2dwire.go by fixportal2, cmd/d2drill/live.go (one line) by m1-liveproof (cmd/d2drill/**). The edits are small and disjoint from those cards' open tasks; the coordinator merges fixstall2 first.
