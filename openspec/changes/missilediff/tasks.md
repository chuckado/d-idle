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
  - docs/coverage/rule-coverage.md
  - openspec/changes/missilediff/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict; coverage.py --check"
result: "real-build differential: 120,000 cases, 44,456 differ before, 0 after; D1-D3 CONFIRMED; local fork branch local/missilediff-record; offline, not verified live"
live: "docs only, nothing run live"
```

- [x] d2d scratch build with the test mode (worker host, fork worktree)
- [x] game.exe side extended (owner stat rewrite, levels 1..40, 120,000 cases)
- [x] before / after counts
- [x] local fix branch, trace doc, ledger rows, merge lines
