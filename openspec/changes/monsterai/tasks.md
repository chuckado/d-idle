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
result: "pending"
live: "docs and emulator only, nothing run live"
```

- [ ] per-AI table, Q1.. with tags
- [ ] oracle runs A and B, counts in the doc
- [ ] deviations D1.., bot findings B1.., drill recipe, matrix proposals
- [ ] ledger rows, merge-lines.txt
