# Tasks

```yaml
change: fixstatic
status: in-progress
agent: fixstatic
branch: agent/fixstatic
worktree: ~/code/d2lod-wt/fixstatic
skip_specs: true
owns:
  - internal/d2dsoakstat/minute.go
  - internal/d2dsoakstat/metrics.go
  - docs/fixstatic.md
exclusive: []
gates:
  - "go vet ./... and go test -short -count=1 ./... on the worker host; changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result: offline only, not verified live
```

## 1. Work

- [x] 1.1 confirm the opener, fallback and Andariel script are inert without the learned skill
- [x] 1.2 `table=` on the minute line; d2dsoakstat and d2dsoak
- [x] 1.3 tests (no Static Field without the skill, Andariel kite, table parse and log)
- [x] 1.4 docs/fixstatic.md, combat.md, BACKLOG, INDEX, CHANGELOG
