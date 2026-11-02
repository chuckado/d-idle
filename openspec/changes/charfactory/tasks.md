# Tasks

```yaml
change: charfactory
status: in-progress
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
result: "docs only; read-only worker-host look (directory listing, .d2s header read, ledger count); nothing run live"
live: "not run"
```

- [x] Inventory, capability table, snapshot library plan, ranked backlog
