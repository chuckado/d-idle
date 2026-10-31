# Tasks

```yaml
change: fix-nav
status: archived
agent: fixnav
branch: agent/fixnav
worktree: ~/code/d2lod-wt/fixnav
owns:
  - internal/nav/walk.go
  - internal/nav/nav.go
  - internal/nav/waypoint.go
  - internal/nav/act2hints.go
  - internal/nav/fixnav_test.go
  - internal/bot/waypoints.go
  - internal/bot/task.go
  - internal/bot/fixnav_test.go
  - docs/fixnav.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... plus go test -race ./internal/nav ./internal/bot ./internal/d2dworld pass on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof is a drill after merge: DW variants in docs/fixnav.md"
merge: 7261d23
result: "offline plus live-proved 2026-10-31 DRILLS5 (see live: line); mouth bounce turn, touched-waypoint drain, towns 75/103/109 level-relative positions; not verified live. m2-tripwp (deferred) released internal/bot/waypoints.go to this card."
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-fnav.txt, docs/probes/drills/2026-10-31-drills5-dw-a9.txt, -a10, -a11): DW.m1 PASS (standing in the Cave Level 1 mouth trigger of Cold Plains, waypoint to town: 1 bounce into the cave, the next step-away went the other way, town in 497 ticks), DW.m2 PASS (0 bounces, 554 ticks), DW.a12 PASS (Cold Plains waypoint touched in passing, DrainActivated [3], a later trip to 3 works), DW.a9/a10/a11 PASS run alone in a fresh process (towns 75, 103, 109 never visited: no none-on-level error); each twice, identical"
```

- [x] 1.1 mouth step turns after a bounce
- [x] 1.2 touched waypoints drain into the runner
- [x] 1.3 level-relative waypoint positions for towns 75, 103, 109
- [x] 1.4 tests, spec delta, docs/fixnav.md
- [x] 1.5 live proof: drill variants (not run here)
