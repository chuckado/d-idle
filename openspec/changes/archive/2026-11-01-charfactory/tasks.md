# Tasks

```yaml
change: charfactory
status: archived
agent: charfactory
branch: agent/charfactory
worktree: ~/code/d2lod-wt/charfactory
skip_specs: true
owns:
  - docs/research/character-factory.md
  - openspec/changes/charfactory/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh"
result: "docs only; nothing started; ADR 0014 cited from the unmerged memrules branch"
live: "not run"
merge: f50e477
```

- [x] Inventory, capability table, snapshot library plan, ranked backlog
