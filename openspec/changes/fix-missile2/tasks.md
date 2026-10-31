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
live: "not verified live: offline; real rows taken from the fork's live vetting (docs/forkcheats.md)"
```

- [x] parser reads the real columns (extras optional)
- [x] speed > 0 filter in FillMissiles
- [x] tests with real rows, err answer, dodge decision
- [x] doc line and spec delta
