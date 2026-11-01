# Tasks

```yaml
change: gate1
status: in-progress
agent: gate1
branch: agent/gate1
worktree: ~/code/d2lod-wt/gate1
skip_specs: true
owns:
  - docs/soaks/gate1-rerun.md
  - docs/probes/soak-logs/gate1-*
  - docs/g1-log.md
  - openspec/changes/gate1/**
exclusive:
  - instance:gate1-slot-0
gates:
result: pending
```

- [ ] 1. Soak run on a hidden d2d slot with one injected fault
- [ ] 2. Judge against the Gate 1 criteria, write the entry
