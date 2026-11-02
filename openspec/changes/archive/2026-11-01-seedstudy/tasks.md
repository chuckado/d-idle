# Tasks

```yaml
change: seedstudy
status: archived
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
result: "docs only; the pin is untested live; a read-only SEEDPIN probe settles the orig-spawns conflict"
live: "docs only, nothing run live"
merge: 47920c1
```

- [x] seed choice on both sides, pinning, RNG advance
- [x] observation and input parity, P3 plan, verdict, first-divergence list, ranked cards
- [x] merge-lines.txt, checks
