# Tasks

```yaml
change: fix152win
status: in-progress
agent: fix152win
branch: agent/fix152win
worktree: ~/code/d2lod-wt/fix152win
owns:
  - docs/research/pr-price-evidence-pack.md
  - openspec/changes/fix152win/**
exclusive:
  - upstream-pr:price
gates:
  - "test_rules green on the worker host, pr_guard 0 FAIL 0 WARN against upstream/main"
result: "not verified on MSVC (no Windows build here); test_rules ok on the worker host; price oracle not re-run (needs the game install, not mounted) but no price code changed"
```

## 1. Work

- [x] 1.1 blame: line 67 ours (95557754), line 509 upstream
- [x] 1.2 rename our local, test_rules, pr_guard
- [x] 1.3 push to the PR branch, evidence pack posted line
