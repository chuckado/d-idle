# Tasks

```yaml
change: mercstats
status: in-progress
agent: mercstats
branch: agent/mercstats
worktree: ~/code/d2lod-wt/mercstats
skip_specs: true
owns:
  - docs/research/mercstats-trace.md
  - tools/emu/mst_*
  - docs/coverage/ledger.tsv
  - docs/coverage/rule-coverage.md
  - openspec/changes/mercstats/**
exclusive: []
gates:
  - tools/emu/mst_cmp.py reports 0 mismatches for each differential named in the trace
  - go run ./cmd/changecheck and go run ./cmd/leakcheck -mode fail clean
result: trace written; V run on the worker host (emulator, d2d scratch build), not verified live
```

- [x] 1 Read AGENTS.md, the pets-merc and monster-stats traces, upstream hireling / merc-ai notes
- [x] 2 Disassemble the level-up, offer, price, XP, potion and resist functions (read)
- [x] 3 Differentials: levelup 2,940, hire 108,000, revive 256, think 100,000, attack 100,000, XP 100,000, potion filter 659 classes
- [x] 4 Trace doc with Q1-Q7, D1-D4, B1-B6, drill recipe, matrix text
- [x] 5 Ledger row and rule-coverage regenerated
- [x] 6 Card, OWNERS, merge-lines; changecheck, leakcheck
