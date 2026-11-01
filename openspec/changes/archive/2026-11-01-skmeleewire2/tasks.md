# Tasks

```yaml
change: skmeleewire2
status: archived
agent: skmeleewire2
branch: agent/skmeleewire2
worktree: ~/code/d2lod-wt/skmeleewire2
skip_specs: true
owns:
  - internal/d2dsrc/gearstats.go
  - internal/d2dsrc/gearstats_test.go
  - docs/skmeleewire2.md
  - openspec/changes/skmeleewire2/**
exclusive: []
gates:
  - "go test -p 1 ./internal/d2dsrc ./internal/combat ./internal/resource"
result: "offline only, table tests pass; not verified live (live checklist in docs/skmeleewire2.md, needs the stats binary); no gear key means defaults unchanged"
live: "offline only, nothing run live"
merge: 47970bf
```

## 1. Work
- [x] 1.1 Parse gear / weapons from dump stats into Player.Stats
- [x] 1.2 Table-driven test
- [x] 1.3 docs/skmeleewire2.md
