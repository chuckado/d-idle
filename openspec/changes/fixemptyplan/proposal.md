## Why

Gate 1 soak 2 logged `combat move failed: no path: empty path` 1,191 times (1,185 in the combat path), 90 percent from the out-of-mana back-off (docs/probes/postfix-soak2-analysis.md, F4, VERIFIED). `empty path` is a planner result: a blocked goal snaps onto the player's own cell, `plan` returns no waypoints and no error, `PathTo` returned nil with a nil error (so `Fighter.valid` accepted the point), and `GoTo` then failed on the zero-length path. Cost is small (about 71 s); the fix removes log noise, wasted ticks and a wrong failure signal.

## What Changes

- internal/input/client.go: `ErrAlreadyThere`, the shared typed "arrived" result.
- internal/nav: `PathTo` and `walk` return an error matching both `ErrAlreadyThere` and `ErrNoPath` (older callers unchanged) for an empty plan; `exit.go` matches the typed error instead of the error text.
- internal/combat: `valid` (back-off) rejects an already-there point so `awayPoint` takes its next candidate (bounded: 3 rings x 10 angles) or the caller falls through to its next option; `validApproach` (firing positions) and `assess` treat it as an arrival; `move` treats it as a quiet tick, not a failed walk.
- docs/fixemptyplan.md, primitives note, BACKLOG, INDEX, CHANGELOG. No numeric gate changes.

Overlap (declared): internal/nav/exit.go is owned by the open card m2-roam (one condition: typed error instead of a text match, same behaviour); docs/primitives.md is owned by m1-liveproof (one P22 note, as fixmeter did for P28). Other touched files are standing `main` defaults.
