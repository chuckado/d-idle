## Why

DRILLS8 finding F13 (docs/probes/drills/README.md): H.pause still fails 3 of 6 first-run tries. With fixstall3 the clearance-0 route frees the stall, but only after about 700 ticks of stalls and side steps, and the budget ends 3 cells short. Two routes were planned and the first did not free the player; the error text reported the previous poll's position.

## What Changes

- internal/d2dctl/approach_path.go: waypoints of a route follow line of sight (a leg is a straight `cmd move`, so it may not cut the corner of an object between two waypoints), a pocket start (isolated free subtile) plans from the nearest free subtile.
- internal/d2dctl/town_verbs.go: ApproachNPC routes at every stall (the blind side step is only the fallback), detects a stop in the middle of a round, grows the budget from the planned route length, reports the final position in the error.
- tests on the real camp grid and a fake world (internal/d2dctl/fixstall4_test.go); a stub grid reply in approach_path_test.go had the `ok` line first (an empty grid), fixed.
- No numeric gate changes (no delta spec).

Overlap (declared, not edited): internal/d2dctl/approach_path.go, approach_path_test.go, town_verbs.go were last changed by the archived fixstall2 and the open fixstall3 (its owns list names only camp-grid.txt, docs/fixstall3.md, wiring_test.go, fight.go, so no overlap with its owned paths); docs/OWNERS has no line for the d2dctl files.
