# Tasks

```yaml
change: fix-nav
status: in-progress
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
merge:
result: "offline only; mouth bounce turn, touched-waypoint drain, towns 75/103/109 level-relative positions; not verified live. m2-tripwp (deferred) released internal/bot/waypoints.go to this card."
live: ""
```

- [x] 1.1 mouth step turns after a bounce
- [x] 1.2 touched waypoints drain into the runner
- [x] 1.3 level-relative waypoint positions for towns 75, 103, 109
- [x] 1.4 tests, spec delta, docs/fixnav.md
- [ ] 1.5 live proof: drill variants (not run here)
