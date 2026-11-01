# Tasks

```yaml
change: fixbuff
status: in-progress
agent: fixbuff
branch: agent/fixbuff
worktree: ~/code/d2lod-wt/fixbuff
skip_specs: true
owns:
  - docs/fixbuff.md
  - internal/combat/fight.go
  - internal/combat/fixbuff_test.go
  - internal/d2/states.go
  - internal/d2dctl/verbs.go
  - internal/d2dctl/parse.go
  - internal/d2dctl/parse_test.go
  - internal/d2dsrc/source.go
  - internal/d2dsrc/source_test.go
  - internal/d2denv/env.go
exclusive: []
gates:
  - "go test -p 1 ./internal/combat ./internal/d2 ./internal/d2dctl ./internal/d2dsrc ./internal/d2denv passes including TestFightBuffRecastsInLongFight and TestBuffStates"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: ""
live: "not verified live: see the live proof recipe in docs/fixbuff.md"
```

- [x] B1 fight-loop buff check (fightBuff)
- [x] B2 armor state ids and buffStates
- [x] B3 Player.States from bare `debug states` (parser unverified live)
- [x] tests (600 s fights, levels 1 and 2, clock and state, surrounded, refused verb)
- [x] docs/fixbuff.md
- [ ] live proof (rerun the DRILLS13 armor recast proof as a long fight)
