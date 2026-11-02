# Tasks

```yaml
change: parityaudit
status: archived
agent: parityaudit
branch: agent/parityaudit
worktree: ~/code/d2lod-wt/parityaudit
skip_specs: true
owns:
  - docs/research/backend-parity.md
  - openspec/changes/parityaudit/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs only; PRESENT-TESTED for the memory backend rests on the g1-log, not on a run today"
live: "docs only, nothing run live"
merge: f1fdda0
```

- [x] build and test state (this checkout and the worker host)
- [x] feature parity matrix
- [x] gate tooling gaps and cost estimate
- [x] smoke runbook and ordered backlog
