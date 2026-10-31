# Tasks

```yaml
change: bundleclick
status: in-progress
agent: bundleclick
branch: agent/bundleclick
worktree: ~/code/d2lod-wt/bundleclick
skip_specs: true
owns:
  - docs/research/pr-click-evidence-pack.md
  - docs/research/pr-tome-evidence-pack.md
  - docs/research/pr-bundle-queue.md
  - docs/decomp-queue.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "changecheck 0 warnings; leakcheck fail mode exit 0"
result: "PR #149 now holds both claims (5 commits, 19 files, head ecb267db, plain push); click oracle 129,283 -> 0 of 390,000 and tome oracle 29,551 -> 0 of 108,000 re-proved on the combined branch; verified live: not applicable (no game)"
```

## 1. Work

- [x] 1.1 Cherry-pick the click commits onto the #149 head, re-prove, push, edit the PR
- [x] 1.2 Log it in the docs
