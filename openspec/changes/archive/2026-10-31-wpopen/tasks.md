# Tasks

```yaml
change: wpopen
status: archived
agent: wpopen
branch: agent/wpopen
worktree: ~/code/d2lod-wt/wpopen
skip_specs: true
merge: 0e592a6
owns:
  - docs/research/pr-wp-evidence-pack.md
  - docs/research/pr-bundle-queue.md
  - docs/BACKLOG.md
  - openspec/changes/wpopen/**
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: "PR opened: psi29a/d2decomped#153 (9 files, 2 commits, tooling 890adf1d + claim 44353293) on upstream main 72132b99. Replay clean; test_game, test_combat, test_rules, test_protocol, pad_use_d2d pass; C++ harness 0 differences of 140,002 single-player cases; pr_guard 0/0; presubmit --quick 0 FAIL; MSVC sweep found one C4456 (unit test `camp` hid main's `camp`), fixed in the claim commit. Not verified live (offline)."
```

## 1. Work

- [x] 1.1 Replay on upstream main, re-run tests, harness, pr_guard, presubmit
- [x] 1.2 MSVC sweep of the diff; fix
- [x] 1.3 Push the branch to the fork; open the PR
- [x] 1.4 posted lines, queue row, BACKLOG, OWNERS
