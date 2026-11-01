# Tasks

```yaml
change: soakg1d2d
status: in-progress
agent: soakg1d2d
branch: agent/soakg1d2d
worktree: ~/code/d2lod-wt/soakg1d2d
skip_specs: true
merge: pending
owns:
  - docs/probes/postfix-soak2.md
  - docs/probes/soak-logs/postfix2-minutes.jsonl.gz
  - docs/review-strategy.md
  - openspec/changes/soakg1d2d/**
exclusive: []
gates:
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0"
result: "pending"
```

## 1. Work

- [x] 1.1 Soak 2 run and report
- [ ] 1.2 Merge and archive
