# Tasks

```yaml
change: verifymissile
status: in-progress
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
result: "Docs only: docs/research/pr-missilerecord-verifier.md. Verdict NOT READY; nothing posted upstream."
live: "docs only, nothing run live"
```

- [x] seed choice on both sides, pinning, RNG advance
- [x] observation and input parity, P3 plan, verdict, first-divergence list, ranked cards
- [x] merge-lines.txt, checks
