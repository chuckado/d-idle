# Tasks

```yaml
change: fix-loot
status: in-progress
agent: fixloot
branch: agent/fixloot
worktree: ~/code/d2lod-wt/fixloot
owns:
  - cmd/d2bot/adapters.go
  - cmd/d2bot/behaviors.go
  - cmd/d2bot/looteradapter_test.go
exclusive: []
gates:
  - "go test -p 1 ./cmd/d2bot passes including TestRecoveryPickupThroughTheAdapter (pile and sword picked only while recovering, totals non-zero); the build fails if looterAdapter drops SetRecovering, RecoveryTotals, LeftBehind, ClearLeftBehind or EvaluateUnit"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge:
result: "offline only, NOT verified live. Fix shared by both backends. Recovery now collects gold piles and sellables (intended design, docs/recovery.md): gold economy of recovery changes, not measured; verify with a soak that `recovery: loot` lines are non-zero. The scenario test needs the game tables (skips without them)."
```

## 1. Work

- [x] 1.1 Forward SetRecovering and RecoveryTotals on looterAdapter; compile-time assertions; scenario test through the adapter
