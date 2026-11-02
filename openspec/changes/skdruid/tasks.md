# Tasks

```yaml
change: skdruid
status: in-progress
agent: skdruid
branch: agent/skdruid
worktree: d2lod-wt/skdruid
skip_specs: true
owns:
  - openspec/changes/skdruid/**
  - docs/research/skdruid-trace.md
  - tools/emu/skdruid_calc.py
  - tools/emu/skdruid_test.cpp
  - tools/emu/skdruid_run.sh
exclusive: []
gates:
  - "changecheck 0 warnings; leakcheck -mode fail 0"
  - "ledger rows only with case counts"
result: "Offline, not verified live. Calc aspect of Skills 221..250: game.exe FUN_00646ca0 vs d2d eval_calc, 0 differ of 27,800 cases (seeds 1 and 7). Effects not run; no deviation, no PR candidate."
credit:
  author: owner
  agents: [Claude Code]
```

## 1. Work

- [x] 1.1 Calc differential for the 30 Druid rows
- [x] 1.2 Trace doc, ledger row, coverage.py --write
- [ ] 1.3 Second pass: effects (forms, Twister / Tornado, Hurricane, Armageddon, summon spawn)
