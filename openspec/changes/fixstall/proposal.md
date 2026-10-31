## Why

DRILLS5 (docs/probes/drills/README.md) found two stalls: F1 H.pause, the walk to Kashya from the camp arrival cell took 396 to 950 ticks (16 to 38 s) against 15 s and 20 s timeouts; F2 D6.d, a pinned `GoTo` climbed the whole ladder (66 s) and ended in ErrStuck without ever using the town portal, which only the exit path had.

## What Changes

- internal/d2dctl town_verbs.go: `ApproachNPC` detects a stall after about 1.2 s (was a whole 3 s round) and its game-time budget is the timeout plus one second per cell of the straight line to the NPC.
- internal/d2dtown adapter.go: the wall limit of the NPC Interact call is 90 s (the budget sizes itself).
- internal/nav walk.go, nav.go: a plain `GoTo`/`GoToUnit` that is stalled in a pocket (3 rungs, still within 3 subtiles of the first stall) runs the Escape hook (town portal) once, ending the walk with ErrLevelChanged; the same before ErrStuck; the stall timer is halved for the rungs after the first when nothing moves. Exit and explore walks are unchanged.
- No numeric gate of a spec changes (no delta).
