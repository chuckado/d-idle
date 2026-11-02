# Tasks

```yaml
change: baselineaudit
status: in-progress
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
result: "Docs only; numbers recomputed from the committed baseline files and logs of the repo; nothing run on the game host, not verified live."
live: "not run"
```

- [x] read specs, baseline files, tools, g1-log, d2d docs
- [x] recompute spread, confidence intervals and power
- [x] write docs/research/baseline-audit.md
- [x] gates (changecheck, leakcheck, ci check)
