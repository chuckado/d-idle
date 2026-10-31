# Tasks

```yaml
change: fix153lint
status: in-progress
agent: fix153lint
branch: agent/fix153lint
worktree: ~/code/d2lod-wt/fix153lint
skip_specs: true
owns:
  - openspec/changes/fix153lint/**
  - docs/research/pr-wp-evidence-pack.md
  - docs/BACKLOG.md
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: "fixup 51c960e5 pushed to the PR branch as a fast-forward; CI re-run states in the report. Verified on the worker host: clang-tidy names-and-includes on the touched files, test_game with game data, test_rules, test_combat, test_protocol, pad_use_d2d, pr_guard 0 FAIL / 0 WARN, -Wshadow sweep."
```

## 1. Record

- [x] 1.1 Evidence pack posted line
- [x] 1.2 BACKLOG lesson under REFOCUS
