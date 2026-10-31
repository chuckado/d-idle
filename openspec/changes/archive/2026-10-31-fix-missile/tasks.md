# Tasks

```yaml
change: fix-missile
status: archived
agent: fixmissile
branch: agent/fixmissile
worktree: ~/code/d2lod-wt/fixmissile
merge: 77188c0
owns:
  - internal/d2dctl/verbs.go
  - internal/d2denv/reader.go
  - internal/env/tuning_combat.go
  - internal/env/tuning_combat_test.go
  - internal/combat/sim_test.go
  - internal/combat/reach_test.go
  - internal/combat/castfirst_test.go
  - docs/knobs.md
exclusive: []
gates:
  - "go test -p 1 ./internal/d2dsrc ./internal/d2dctl ./internal/combat ./internal/env passes including TestMissiles*, TestCastGapCoversEarlyCast, TestHostileMissilesD2dShape"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "verified live 2026-10-31 DRILLS5; see the live: line"
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-ffight.txt): D5.dodge PASS (Fallen Shamans spawned near a Sorceress with 400 life: 8 dodges and 1.00 cast commands per cast over 17 casts in one run; other runs 4 to 17 dodges; the idle sampling sees the missiles in debug missiles and in Snapshot.Missiles); commands per cast stay at or below 1.0 in every run (CastGapMs 650)"
```

- [x] B1: d2d has no missile listing; verb request written; parser, source mapping, latch, fake-source tests
- [x] B3: CastGapMs 650, evidence in knob comment and docs
- [x] spec deltas
- [x] fork agent adds `debug missiles`; live proof: drill D5 rerun (m1-liveproof)

Note: owns lines for parse.go, missiles.go, missiles_test.go, dodge_test.go and docs/fixmissile.md released to fix-missile2 (its code is merged).
