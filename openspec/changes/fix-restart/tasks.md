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
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-d10.txt, two runs, identical verdict): D10 PASS (kill -9 of the slot d2d under the managed bot: back in game within 8 s, minute line restarts 0 to 1 and 1 to 2 = delta exactly 1, xpRisk after the kill 108 = the unsaved XP only (xp 1863, saved 1755; before the kill 588), loop resumed, saved XP not lower, no leftovers, P35 clean)"
```

- [x] game-lost restart counted in the minute line's restarts
- [x] d2d SavedExp from the save file under the slot's d2d home
- [x] scenario tests and spec delta
- [ ] live proof: drill D10 rerun
