# Tasks

```yaml
change: fix-wp
status: archived
agent: fixwp
branch: agent/fixwp
worktree: ~/code/d2lod-wt/fixwp
owns:
  - internal/nav/waypoint.go
  - internal/nav/wpclick_test.go
  - internal/nav/exit_test.go
  - internal/d2dsrc/items.go
  - internal/d2dsrc/source_test.go
  - docs/fixwp.md
  - internal/d2map/tracker.go
  - internal/d2dworld/world.go
exclusive: []
gates:
  - "go test -short ./internal/nav ./internal/d2dsrc ./internal/drill ./cmd/d2drill passes including TestNearestWaypointUnit, TestWaypointRowCentresHitBoxes, TestWaypointClicksEveryLevel, TestWaypointRowHasWaypointClass, TestDW"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: 3eba209
result: "W1, W2, W5 fixed with scenario tests; verified live (drill group DW, each twice identical): a1 a2 a4 a5 b c PASS, a3 BLOCKED (act unlock cheat); see docs/fixwp.md"
live: "docs/probes/drills/2026-10-31-dw.txt"
```

- [x] W1 waypoint class in d2dsrc
- [x] W2 row and tab centres
- [x] scenario tests and spec delta
- [x] drill group DW
- [x] live proof (docs/fixwp.md)
