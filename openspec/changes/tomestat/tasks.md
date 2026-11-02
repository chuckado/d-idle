# Tasks

```yaml
change: tomestat
status: in-progress
agent: tomestat
branch: agent/tomestat
worktree: d2lod-wt/tomestat
skip_specs: true
owns:
  - docs/tomestat.md
  - internal/d2dsrc/groundid_test.go
  - openspec/changes/tomestat/**
exclusive: []
gates:
  - "go test -p 1 ./internal/d2dsrc ./internal/d2dloot ./internal/d2denv ./internal/d2dtown; changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Ground and carried extras separated; offline tests pass; NOT verified live."
live: "not verified; checklist in docs/tomestat.md"
credit:
  author: owner
  agents: [Claude Code]
```

- [x] root cause from the Gate 2 log and the fork source
- [x] Extras.Ground, d2dloot reads it
- [x] regression test
- [x] doc
