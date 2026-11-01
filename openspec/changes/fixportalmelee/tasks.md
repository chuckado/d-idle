# Tasks

```yaml
change: fixportalmelee
status: in-progress
agent: fixportalmelee
branch: agent/fixportalmelee
worktree: ~/code/d2lod-wt/fixportalmelee
skip_specs: true
owns:
  - docs/fixportalmelee.md
  - internal/combat/safety.go
  - internal/combat/fighter.go
  - internal/combat/crowd.go
  - internal/combat/resource.go
  - internal/combat/portalmelee_test.go
  - internal/resource/types.go
  - internal/resource/abort.go
  - internal/resource/portalmelee_test.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat ./internal/resource passes including TestSafetyMeleeCutCastScenario, TestMeleeCastGuardOrder, TestAbortFightCutCast"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "control PASS x3 (one adjacent: portal cast); D-portalmelee FAIL x3 in DRILLS11 (F18, rule unreachable at radius 3), PASS x3 in DRILLS12 with fixcrowd (crowdAdjacent 8): Safety chicken, scroll kept; a monster hit does not cut a cast in 14 of 14 trials, a debug hit does; README DRILLS11 and DRILLS12"
```

- [x] escape logic found (Safety Decide, resource.AbortFight)
- [x] rule added (drink, else chicken, when 3 or more attackers are adjacent)
- [x] scenario and table tests
- [x] docs, primitives note, backlog, changelog
- [ ] live check (hidden slot)
