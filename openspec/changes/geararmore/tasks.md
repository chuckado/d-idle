# Tasks

```yaml
change: geararmore
status: in-progress
agent: pa-geararmore
branch: agent/geararmore
worktree: ~/code/d2lod-wt/geararmore
skip_specs: true
owns:
  - openspec/changes/geararmore/**
  - docs/research/geararmore.md
  - tools/emu/gar_run.py
  - tools/emu/gar_dis.py
  - tools/emu/gar_census.sh
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "every V run number has a command, a case count and a seed in the doc"
result: "Offline, nothing verified live, no game, no Go change. Stat 179 (FUN_0057d9b0, 100,032 cases) and stat 120 (FUN_0057b7d0 branch, 100,648 cases) run natively on the worker host: 0 differ from the reading. Table census (D): 179 is on no row of the item tables. Bot finding B1: the fork gear sums omit stat 119 and the five monster-specific stats. No PR candidate. Ledger not changed."
```

## 1. Work

- [x] 1.1 Read FUN_0057d9b0 (stat 179 loop) and FUN_0057b7d0 (stat 120) in the disassembly
- [x] 1.2 gar_run.py parts a and b, 100,000 random + boundaries each, seed 7
- [x] 1.3 Table census gar_census.sh
- [x] 1.4 Doc docs/research/geararmore.md, card, OWNERS, merge-lines
