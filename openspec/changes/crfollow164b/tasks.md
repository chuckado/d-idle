# Tasks

```yaml
change: crfollow164b
status: in-progress
skip_specs: true
agent: crfollow164b
branch: agent/crfollow164b
worktree: ~/code/d2lod-wt/crfollow164b
owns:
  - tools/emu/mm_diff.py
  - openspec/changes/crfollow164b/**
exclusive:
  - upstream-pr:monster-missile-record-and-poison
gates:
  - "pr_guard 0 FAIL 0 WARN on the new commit and the range"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: "tooling only; commit 11bc2c4d on the PR branch; not verified live"
```

## 1. Work

- [x] 1.1 read the finding, decide valid, commit, run checks, push, reply
- [x] 1.2 mirror the fix to the public tools/emu/mm_diff.py
