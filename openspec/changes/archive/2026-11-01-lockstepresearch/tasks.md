# Tasks

```yaml
change: lockstepresearch
status: archived
agent: lockstepresearch
branch: agent/lockstepresearch
worktree: ~/code/d2lod-wt/lockstepresearch
skip_specs: true
owns:
  - docs/research/original-game-lockstep.md
  - openspec/changes/lockstepresearch/**
exclusive: []
gates:
  - go run ./cmd/changecheck: 0 warnings
  - go run ./cmd/leakcheck: 0 hits
result: "docs only; nothing started; hook policy conditions listed"
merge: 8961692
```

- [x] 1 prior art table with sources
- [x] 2 feasibility, tiers, probe runbook, policy, recommendation
- [x] 3 merge-lines.txt, checks
