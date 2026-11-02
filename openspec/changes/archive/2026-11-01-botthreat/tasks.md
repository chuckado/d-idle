# Tasks

```yaml
change: botthreat
status: archived
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
result: "offline tests only; runnerGuard default ON, threatDodge/shooterKillMin/standOK default OFF; live drill checklist in docs/botthreat.md; d2d feeds no missiles so dodge does nothing there"
live: "not run live; checklist in docs/botthreat.md"
merge: f8f9666
```

- [x] threat table, verdict, ranking
- [x] kill-first from monstats columns
- [x] runner guard, standing Fallen
- [x] knobs and docs
