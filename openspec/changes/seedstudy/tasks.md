# Tasks

```yaml
change: seedstudy
status: in-progress
agent: seedstudy
branch: agent/seedstudy
worktree: ~/code/d2lod-wt/seedstudy
skip_specs: true
owns:
  - docs/research/seed-study.md
  - openspec/changes/seedstudy/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck 0 hits; tools/ci/check.sh"
result: "Docs only: docs/research/seed-study.md. Nothing run live, nothing built."
live: "docs only, nothing run live"
```

- [x] seed choice on both sides, pinning, RNG advance
- [x] observation and input parity, P3 plan, verdict, first-divergence list, ranked cards
- [x] merge-lines.txt, checks
