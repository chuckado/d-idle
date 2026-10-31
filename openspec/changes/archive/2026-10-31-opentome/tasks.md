# Tasks

```yaml
change: opentome
status: archived
merge: b9579c5
agent: opentome
branch: agent/opentome
worktree: ~/code/d2lod-wt/opentome
skip_specs: true
owns:
  - docs/research/pr-tome-evidence-pack.md
  - docs/research/pr-bundle-queue.md
  - docs/research/README.md
  - docs/decomp-queue.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "changecheck 0 warnings; leakcheck fail mode exit 0"
result: "PR #149 opened on psi29a/d2decomped (3 commits, 8 files) after the sync onto upstream 0dc4c62a; oracle 29,551 and 56,616 -> 0 and 0 re-proved; CodeRabbit started by itself, no trigger posted; verified live: not applicable (no game)"
```

## 1. Work

- [x] 1.1 Sync, re-prove, push, open the PR
- [x] 1.2 Log it in the docs
