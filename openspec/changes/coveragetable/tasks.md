# Tasks

```yaml
change: coveragetable
status: in-progress
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
result: ""
live: "docs and tooling only, nothing run live"
```

- [ ] generator, denominators, ledger seeded from the docs
- [ ] tests, README, pointer line in decomp-queue
- [ ] checks and merge-lines
