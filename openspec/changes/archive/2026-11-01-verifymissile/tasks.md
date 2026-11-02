# Tasks

```yaml
change: verifymissile
status: archived
agent: verifymissile
branch: agent/verifymissile
worktree: ~/code/d2lod-wt/verifymissile
skip_specs: true
owns:
  - docs/research/pr-missilerecord-verifier.md
  - openspec/changes/verifymissile/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck 0 hits; tools/ci/check.sh"
result: "docs only; changes required listed; Andariel poison bolt has the same bug on upstream"
live: "docs only, nothing run live"
merge: af74c4d
```

- [x] seed choice on both sides, pinning, RNG advance
- [x] observation and input parity, P3 plan, verdict, first-divergence list, ranked cards
- [x] merge-lines.txt, checks
