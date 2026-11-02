# Tasks

```yaml
change: kitgatelevel
status: in-progress
agent: kitgatelevel
branch: agent/kitgatelevel
worktree: ~/code/d2lod-wt/kitgatelevel
skip_specs: true
owns:
  - docs/kitgate.md
  - docs/knobs.md
  - internal/bot/kitgate.go
  - internal/bot/kitgate_test.go
  - internal/bot/kitlevel_test.go
  - internal/bot/observe.go
  - internal/bot/resource.go
  - internal/bot/goal.go
  - internal/resource/tuning.go
  - internal/env/tuning_resource.go
  - openspec/changes/kitgatelevel/**
exclusive: []
gates:
  - go test -p 1 ./internal/bot ./internal/resource ./internal/env ./tools/knobdump passes
result: offline only (mem fakes); NOT verified live.
```

- [x] knobs and the one kit definition
- [x] kitTrip, kitHolds, cap and kitgate-skip log
- [x] table-driven tests
- [x] docs/kitgate.md rules and live drill checklist, docs/knobs.md regenerated
- [ ] live drill (gate 1 rerun): see docs/kitgate.md
