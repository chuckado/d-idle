# Tasks

```yaml
change: d2d-gate-ladder
status: in-progress
skip_specs: false
agent: gateladder
branch: agent/gateladder
worktree: ~/code/d2lod-wt/gateladder
owns:
  - openspec/changes/d2d-gate-ladder/**
  - docs/review-strategy.md
  - docs/adr/0013-retire-game-exe-gates.md
  - docs/decisions.md
  - docs/g1-log.md
  - docs/process/dispatcher-brief.md
exclusive: []
gates:
  - "tools/openspec.sh validate --all --strict passes with this delta"
  - "go run ./cmd/leakcheck -mode fail has 0 hits; changecheck -cli 0 warnings"
merge:
result:
```

## 1. Work

- [x] 1.1 ADR 0013 and the decisions.md index line
- [x] 1.2 delta spec for soak-gates (Gate 0 to 3, G1 removed, G2 and start guard modified)
- [x] 1.3 review-strategy banner and d2d gate ladder subsection; g1-log banner
- [x] 1.4 proposed dispatcher step 4 (not live until the user says go)
- [x] 1.5 candidate drill D10 proposed in the report (not built)
- [ ] 1.6 apply the spec delta by hand at archive (coordinator)
