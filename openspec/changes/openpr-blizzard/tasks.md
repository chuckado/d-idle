# Tasks

```yaml
change: openpr-blizzard
status: in-progress
agent: openpr-blizzard
branch: agent/openpr-blizzard
worktree: ~/code/d2lod-wt/openpr-blizzard
skip_specs: true
owns:
  - openspec/changes/openpr-blizzard/**
  - docs/research/openpr-log.md
exclusive:
  - "upstream-pr:blizzard-shard-offsets"
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "Opened upstream #163 (Blizzard and Eruption shard offsets; 5 files, 2 commits) from the fork branch fix/blizzard-shard-offsets-v2, plain push, nothing else posted. Checks re-run on the worker host (pr_guard, Release build, test_game, test_rules, both differentials, ruff, presubmit --quick, merge-tree). Verified offline, not live."
```

## 1. Work

- [x] 1.1 Fetch upstream main (7d5b3a5c, unchanged), re-vet the candidate head 3d6f3fe7
- [x] 1.2 Push the new fork branch (plain) and open #163 (body ends `Agent: Claude Code`)
- [x] 1.3 Posted log and queue row
