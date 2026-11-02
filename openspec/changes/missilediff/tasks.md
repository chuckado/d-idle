# Tasks

```yaml
change: missilediff
status: in-progress
agent: missilediff
branch: agent/missilediff
worktree: ~/code/d2lod-wt/missilediff
skip_specs: true
owns:
  - docs/research/missilediff.md
  - tools/emu/mm_diff.py
  - tools/emu/mm_diff_run.sh
  - docs/coverage/ledger.tsv
  - openspec/changes/missilediff/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict; coverage.py --check"
result: ""
live: "docs only, nothing run live"
```

- [ ] d2d scratch build with the test mode (worker host, fork worktree)
- [ ] game.exe side extended (owner stat rewrite, levels 1..40, 120,000 cases)
- [ ] before / after counts
- [ ] local fix branch, trace doc, ledger rows, merge lines
