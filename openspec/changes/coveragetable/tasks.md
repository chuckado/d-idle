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
result: "docs/coverage/rule-coverage.md generated (21 subsystems, 4910 rows, ledger seeded from the research docs), coverage.py --check wired into tools/ci/scripts.sh, 12 unit tests. Not verified live (docs and tooling only)."
live: "docs and tooling only, nothing run live"
```

- [x] generator, denominators, ledger seeded from the docs
- [x] tests, README, pointer line in decomp-queue
- [x] checks and merge-lines
