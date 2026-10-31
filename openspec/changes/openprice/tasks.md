# Tasks

```yaml
change: openprice
status: in-progress
skip_specs: true
agent: openprice
branch: agent/openprice
worktree: ~/code/d2lod-wt/openprice
owns:
  - docs/research/pr-price-evidence-pack.md
  - docs/decomp-queue.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/openprice/**
exclusive:
  - upstream-pr:price
gates:
  - "pr_guard 0 FAIL 0 WARN, 8 files, upstream not moved"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: "PR opened psi29a/d2decomped #152 from fix/vendor-prices (1bfff15c), no replay needed; our repo not pushed."
```

## 1. Work

- [x] 1.1 fetch upstream, pr_guard, push to the fork, open the PR
- [x] 1.2 docs log lines
