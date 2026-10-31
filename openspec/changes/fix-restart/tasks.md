# Tasks

```yaml
change: fix-restart
status: in-progress
agent: fixrestart
branch: agent/fixrestart
worktree: ~/code/d2lod-wt/fixrestart
owns:
  - internal/bot/loop.go
  - internal/bot/observe.go
  - internal/bot/fixrestart_test.go
  - internal/bot/botsmall_test.go
  - cmd/d2bot/d2dbackend.go
  - cmd/d2bot/savedexp.go
  - cmd/d2bot/savedexp_test.go
  - docs/fixrestart.md
exclusive: []
gates:
  - "go test -short ./internal/bot ./cmd/d2bot passes including TestGameLostRestartCountsOnceWithoutMenus, TestXPRiskIsSinceLastSaveAcrossAKill, TestD2DSavedExp*"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "not verified live: rerun drill D10 (docs/fixrestart.md)"
```

- [x] game-lost restart counted in the minute line's restarts
- [x] d2d SavedExp from the save file under the slot's d2d home
- [x] scenario tests and spec delta
- [ ] live proof: drill D10 rerun
