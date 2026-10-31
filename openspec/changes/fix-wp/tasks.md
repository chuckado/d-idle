# Tasks

```yaml
change: fix-wp
status: in-progress
agent: fixwp
branch: agent/fixwp
worktree: ~/code/d2lod-wt/fixwp
owns:
  - internal/nav/waypoint.go
  - internal/nav/wpclick_test.go
  - internal/nav/exit_test.go
  - internal/d2dsrc/items.go
  - internal/d2dsrc/source_test.go
  - internal/drill/d3.go
  - internal/drill/d3_test.go
  - cmd/d2drill/wp_lab.go
  - cmd/d2drill/main.go
  - cmd/d2drill/live.go
  - cmd/d2drill/portal_lab.go
  - docs/fixwp.md
exclusive: []
gates:
  - "go test -short ./internal/nav ./internal/d2dsrc ./internal/drill ./cmd/d2drill passes including TestNearestWaypointUnit, TestWaypointRowCentresHitBoxes, TestWaypointClicksEveryLevel, TestWaypointRowHasWaypointClass, TestD3"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: ""
```

- [x] W1 waypoint class in d2dsrc
- [x] W2 row and tab centres
- [x] scenario tests and spec delta
- [x] drill group D3
- [ ] live proof (docs/fixwp.md)
