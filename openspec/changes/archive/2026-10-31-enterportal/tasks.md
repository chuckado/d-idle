# Tasks

```yaml
change: enterportal
status: archived
agent: enterportal
branch: agent/enterportal
worktree: ~/code/d2lod-wt/enterportal
owns:
  - internal/d2dctl/verbs.go
  - internal/d2dctl/town_verbs.go
  - internal/d2dctl/town_verbs_test.go
  - internal/d2dinput/controller.go
  - internal/d2dinput/controller_test.go
  - internal/d2dsrc/items.go
  - internal/d2dsrc/source.go
  - internal/input/semantic.go
  - internal/town/trip.go
  - internal/town/enterportal_test.go
  - docs/d2d-portal-entry.md
  - docs/d2d-town.md
  - docs/probes/enterportal/**
exclusive: []
gates:
  - "go test ./internal/town (on the worker host, game tables) passes TestScenario_OpenPortalBesideThePlayer; go test ./internal/d2dctl ./internal/d2dinput pass TestEnterPortalSendsInteract and TestUsePortal"
  - "drills D2.t, D2.g, D2.f PASS live on a lockstep lab slot (docs/probes/enterportal/)"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
merge: c3ef097
result: "verified live (lockstep lab slot, 2 identical runs): D2.t PASS (ToTown enters the open portal in 5 ticks, no scroll), D2.g PASS, D2.f PASS (arrival 6 subtiles from origin); probes: hover+click without a drawn frame walks, with step render enters in 1 tick, cmd interact -2000 enters in 1 tick (34 from 4.5 cells), walking onto the cell never enters. Cause: lab lockstep draws no frame, d2d hover pick is computed in the draw (lab artefact, not a d2d deviation)."
```

- [x] 1 find why no portal is entered on d2d (code reading: d2d client hover pick, lockstep draw skip)
- [x] 2 TakePortal verb; EnterPortal and UsePortal use it; d2dsrc portal ids by which
- [x] 3 ToTown: no cast while an open portal is in reach; walk fallback when only it was tried
- [x] 4 offline tests and spec deltas
- [x] 5 live: probe sequences (walk, click, rendered click, interact) and drills D2.t, D2.g, D2.f
- [x] 6 docs/d2d-portal-entry.md, changecheck, leakcheck, reap
