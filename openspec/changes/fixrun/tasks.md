# Tasks

```yaml
change: fixrun
status: in-progress
agent: fixrun
branch: agent/fixrun
worktree: ~/code/d2lod-wt/fixrun
skip_specs: true
owns:
  - internal/d2dinput/run.go
  - internal/d2dinput/run_test.go
  - internal/d2dinput/controller.go
  - internal/d2dinput/lifecycle.go
  - internal/d2denv/runreset_test.go
  - internal/bot/observe.go
  - docs/fixrun.md
  - openspec/changes/fixrun/**
exclusive: []
gates:
  - "go test -p 1 for internal/d2dinput, internal/d2denv, internal/bot passes; whole-repo vet and test -short on the worker host"
  - "live proof (DRILLS lane): restart a d2d game, move a fixed leg, compare ticks per subtile before and after (docs/fixrun.md)"
merge:
result: "offline only: not verified live; the card stays open at merge until the live proof ran"
```

## 1. Work

- [x] 1.1 B1: verify and test the belief reset on every game start (d2denv and d2dinput.Process)
- [x] 1.2 B2: run observables in the minute line (RunInfo, LogKV)
- [x] 1.3 B3: stall wait analysed, left unchanged, test pins it
- [x] 1.4 docs/fixrun.md, primitives row note, BACKLOG, INDEX, CHANGELOG
- [ ] 1.5 live proof (DRILLS lane)
