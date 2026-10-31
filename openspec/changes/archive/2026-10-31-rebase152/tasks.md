# Tasks

```yaml
change: rebase152
status: in-progress
agent: rebase152
branch: agent/rebase152
worktree: ~/code/d2lod-wt/rebase152
skip_specs: true
owns:
  - docs/research/pr-price-evidence-pack.md
  - openspec/changes/rebase152/**
exclusive:
  - upstream-pr:price
gates:
  - "rebase conflicts only in tools/emu/README.md, resolved keeping both sides"
  - "upstream test_rules, test_game, test_combat, test_protocol pass; price harness 0 differ; pr_guard 0 FAIL 0 WARN against upstream/main; presubmit --quick 0 FAIL"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: "rebased and force-pushed with --force-with-lease pinned to a50bdbf3; see the posted: line in the evidence pack. Verified offline on the worker host, not live."
```

## 1. Work

- [x] 1.1 Rebase onto upstream/main, resolve tools/emu/README.md (both sides)
- [x] 1.2 Rebuild, upstream tests, price harness, pr_guard, presubmit --quick, MSVC sweep
- [x] 1.3 Force-push with lease
- [x] 1.4 Evidence pack line, card, OWNERS, changecheck, leakcheck
