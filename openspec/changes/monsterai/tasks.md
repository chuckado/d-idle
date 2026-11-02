# Tasks

```yaml
change: monsterai
status: in-progress
agent: monsterai
branch: agent/monsterai
worktree: ~/code/d2lod-wt/monsterai
skip_specs: true
owns:
  - docs/research/monster-ai-trace.md
  - docs/coverage/ledger.tsv
  - docs/coverage/rule-coverage.md
  - docs/coverage/gaps.tsv
  - tools/emu/mai_*
  - openspec/changes/monsterai/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs and emulator harness: docs/research/monster-ai-trace.md (146-row per-AI table, 25 AIs run twice, 75,000 cases; D1-D11, B1-B3, no PR candidate); ledger monster.ai 11 rows (6 TESTED-0-DIFF, 5 TESTED-DEVIATION). Nothing run live."
live: "docs and emulator only, nothing run live"
```

- [x] per-AI table, Q1.. with tags
- [x] oracle runs A and B, counts in the doc
- [x] deviations D1.., bot findings B1.., drill recipe, matrix proposals
- [x] ledger rows, merge-lines.txt
