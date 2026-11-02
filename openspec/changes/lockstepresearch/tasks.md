# Tasks

```yaml
change: lockstepresearch
status: in-progress
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
result: research doc written from web sources and our docs; nothing run live, nothing built; frame scheduler facts are from D2MOO source, not yet read on our Game.exe
```

- [x] 1 prior art table with sources
- [x] 2 feasibility, tiers, probe runbook, policy, recommendation
- [x] 3 merge-lines.txt, checks
