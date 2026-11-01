# Tasks

```yaml
change: coveragetable
status: archived
agent: coveragetable
branch: agent/coveragetable
worktree: ~/code/d2lod-wt/coveragetable
skip_specs: true
owns:
  - docs/coverage/**
  - tools/coverage/**
  - openspec/changes/coveragetable/**
exclusive: []
gates:
  - "coverage.py --check; changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "docs and tooling only; generator check wired into CI scripts; ledger seeded by Sonnet from trace docs, spot-check owed for id mappings and for the damage and mana rows marked fully run-tested"
live: "docs and tooling only, nothing run live"
merge: 43d606e
```

- [x] generator, denominators, ledger seeded from the docs
- [x] tests, README, pointer line in decomp-queue
- [x] checks and merge-lines
