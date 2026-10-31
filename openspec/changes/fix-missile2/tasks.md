# Tasks

```yaml
change: fix-missile2
status: in-progress
agent: fixmissile2
branch: agent/fixmissile2
worktree: ~/code/d2lod-wt/fixmissile2
owns:
  - internal/d2dctl/parse.go
  - internal/d2dsrc/missiles.go
  - internal/d2dsrc/missiles_test.go
  - internal/combat/dodge_test.go
  - docs/fixmissile.md
exclusive: []
gates:
  - "go test -p 1 ./internal/d2dsrc ./internal/d2dctl ./internal/combat passes including TestMissilesRealVerb, TestParseMissileReal, TestSidestepRealRows"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "verified live 2026-10-31 DRILLS5 (docs/probes/drills/2026-10-31-drills5-ffight.txt): the d2d missile rows reach the bot: with the idle character, debug missiles listed 1 to 2 hostile moving rows and Snapshot.Missiles the same, and the fight loop dodged (D5.dodge PASS, 4 to 17 dodges per run)"
```

- [x] parser reads the real columns (extras optional)
- [x] speed > 0 filter in FillMissiles
- [x] tests with real rows, err answer, dodge decision
- [x] doc line and spec delta
