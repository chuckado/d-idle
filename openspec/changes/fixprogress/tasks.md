# Tasks

```yaml
change: fixprogress
status: in-progress
agent: fixprogress
branch: agent/fixprogress
worktree: ~/code/d2lod-wt/fixprogress
skip_specs: true
owns:
  - docs/fixprogress.md
  - internal/town/chores.go
  - internal/town/town.go
  - internal/town/fixprogress_test.go
  - openspec/changes/fixprogress/**
exclusive: []
gates:
  - "go test -p 1 ./internal/town ./internal/bot passes; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Cause from the Gate 1 rerun log (V loop, U accounting root); fix and 3 sim tests pass on the worker host; NOT verified live, recipe in docs/fixprogress.md"
live: "not run; recipe in docs/fixprogress.md"
```

- [x] timeline and cause from the minute log and the bot log
- [x] town: portal-empty fact and watchdog, tests
- [ ] live proof (later card)
