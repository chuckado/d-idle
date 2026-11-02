# Tasks

```yaml
change: branchtriage
status: in-progress
agent: branchtriage
branch: agent/branchtriage
worktree: d2lod-wt/branchtriage
skip_specs: true
owns:
  - docs/research/branch-triage-20261102.md
  - openspec/changes/branchtriage/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "docs only; read-only triage, nothing run live"
live: "docs only, nothing run live"
```

- [x] recompute the unmerged agent branches against origin/main
- [x] per branch: unique commits, files, content on main by another route, card status, worktree state
- [x] write the triage table with delete / merge commands (not run)
