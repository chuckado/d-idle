# Tasks

```yaml
change: baselineaudit
status: archived
agent: baselineaudit
branch: agent/baselineaudit
worktree: ~/code/d2lod-wt/baselineaudit
skip_specs: true
owns:
  - docs/research/baseline-audit.md
  - openspec/changes/baselineaudit/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh"
result: "docs only; recomputed from the baseline files; raw minute logs not reopened"
live: "not run"
merge: 016bd0b
```

- [x] read specs, baseline files, tools, g1-log, d2d docs
- [x] recompute spread, confidence intervals and power
- [x] write docs/research/baseline-audit.md
- [x] gates (changecheck, leakcheck, ci check)
