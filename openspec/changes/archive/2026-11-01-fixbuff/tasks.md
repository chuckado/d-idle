# Tasks

```yaml
change: fixbuff
status: archived
agent: fixbuff
branch: agent/fixbuff
worktree: ~/code/d2lod-wt/fixbuff
skip_specs: true
owns:
  - docs/fixbuff.md
  - internal/combat/fight.go
  - internal/data/extract.go
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
merge: 27049dd
result: "proven live on d2d: debug states reply parsed, Player.States filled (bit 10 in 66 of 66 samples), armor recast at 115.1 s and 229.2 s in a 300 s fight with 0 absent samples, ids 88/30/20/14 seen; the stale-cache check is offline-only"
live: "verified live (drills16, hidden slot 1, lab build debug-skills): raw `debug states` line parsed by Client.Buffs, Player.States bit 10 set in 12 of 12 up samples and clear after the clear, armor recast from inside a 300 s fight at 115.1 s and 229.2 s with 0 of 66 samples absent, states 20/30/88 and 14 confirmed (R-bs1/R-bs3/R-bs4); stale-cache failure offline only"
```

- [x] B1 fight-loop buff check (fightBuff)
- [x] B2 armor state ids and buffStates
- [x] B3 Player.States from bare `debug states` (parser unverified live)
- [x] tests (600 s fights, levels 1 and 2, clock and state, surrounded, refused verb)
- [x] F1 stale cache: CacheCurrent / Load check required columns
- [x] docs/fixbuff.md
- [x] live proof (drills16: PASS)
