# Tasks

```yaml
change: botthreat
status: in-progress
agent: botthreat
branch: agent/botthreat
worktree: ~/code/d2lod-wt/botthreat
skip_specs: true
owns:
  - internal/combat/threat.go
  - internal/combat/threat_test.go
  - docs/botthreat.md
  - openspec/changes/botthreat/**
exclusive: []
gates:
  - "go test -p 1 ./internal/combat ./internal/env ./tools/knobdump; gofmt; go vet; changecheck 0 warnings; leakcheck 0"
result: "Offline only, mem.Fake / table-driven tests; nothing verified live. runnerGuard ON by default, the rest OFF."
live: "not run live; checklist in docs/botthreat.md"
```

- [x] threat table, verdict, ranking
- [x] kill-first from monstats columns
- [x] runner guard, standing Fallen
- [x] knobs and docs
