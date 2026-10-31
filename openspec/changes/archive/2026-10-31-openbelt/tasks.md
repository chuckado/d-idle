# Tasks

```yaml
change: openbelt
status: archived
merge: 146f748
agent: openbelt
branch: agent/openbelt
worktree: ~/code/d2lod-wt/openbelt
skip_specs: true
owns:
  - docs/research/pr-belt-evidence-pack.md
  - docs/decomp-queue.md
  - docs/BACKLOG.md
  - openspec/changes/openbelt/**
exclusive:
  - upstream-pr:belt
gates:
  - "pr_guard 0/0 and the belt oracle 0 mismatches after the local rebase"
merge: pending
result: "verified live: not applicable (offline decomp PR); PR psi29a/d2decomped#147 opened, head 3963222d"
```

## 1. Work

- [x] 1.1 rebase trace/belt onto upstream/main, re-run the checks
- [x] 1.2 push fix/belt-scroll-columns to the fork, open the PR
- [x] 1.3 posted: lines, queue and backlog
