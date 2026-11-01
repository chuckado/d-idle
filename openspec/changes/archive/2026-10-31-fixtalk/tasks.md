# Tasks

```yaml
change: fixtalk
status: archived
merge: 5f67ea3
agent: fixtalk
branch: agent/fixtalk
worktree: ~/code/d2lod-wt/fixtalk
skip_specs: true
owns:
  - internal/talk/**
  - cmd/d2bot/task.go
  - cmd/d2bot/acts.go
  - cmd/d2bot/acts_test.go
  - cmd/d2bot/cain_test.go
  - docs/fixtalk.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
  - "live proof after merge: D8.q and D8.w PASS with the bot's talk path (recipe in docs/fixtalk.md)"
live: "DRILLS10: D8.q PASS (skill points 0 to 1), D8.w PASS (area 40)"
result: "live PASS (DRILLS10)"
```

## 1. Work

- [x] 1.1 find the talk nodes and the act hop wiring
- [x] 1.2 move tables and Adapter into internal/talk, cmd/d2bot builds and its tests pass
- [x] 1.3 table tests with fakes
- [x] 1.4 D8.q and D8.w call the bot's path
- [x] 1.5 docs, INDEX, CHANGELOG, BACKLOG, primitives note
- [x] 1.6 live proof (DRILLS10)
