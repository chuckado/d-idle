# Tasks

```yaml
change: fix-missile
status: in-progress
agent: fixmissile
branch: agent/fixmissile
worktree: ~/code/d2lod-wt/fixmissile
owns:
  - internal/d2dctl/parse.go
  - internal/d2dctl/verbs.go
  - internal/d2dsrc/missiles.go
  - internal/d2denv/reader.go
  - internal/d2dsrc/missiles_test.go
  - internal/env/tuning_combat.go
  - internal/env/tuning_combat_test.go
  - internal/combat/sim_test.go
  - internal/combat/reach_test.go
  - internal/combat/castfirst_test.go
  - internal/combat/dodge_test.go
  - docs/knobs.md
  - docs/fixmissile.md
exclusive: []
gates:
  - "go test -p 1 ./internal/d2dsrc ./internal/d2dctl ./internal/combat ./internal/env passes including TestMissiles*, TestCastGapCoversEarlyCast, TestHostileMissilesD2dShape"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "not verified live: needs the fork verb `debug missiles`, then drill D5 with a spawned ranged monster (docs/fixmissile.md)"
```

- [x] B1: d2d has no missile listing; verb request written; parser, source mapping, latch, fake-source tests
- [x] B3: CastGapMs 650, evidence in knob comment and docs
- [x] spec deltas
- [ ] fork agent adds `debug missiles`; live proof: drill D5 rerun (m1-liveproof)
