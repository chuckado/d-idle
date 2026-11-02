# Tasks

```yaml
change: skcurseval
status: in-progress
agent: skcurseval
branch: agent/skcurseval
worktree: ~/code/d2lod-wt/skcurseval
skip_specs: true
owns:
  - docs/research/skcurseval-trace.md
  - tools/emu/sk_curseval.py
  - tools/emu/sk_curseval_run.sh
  - tools/emu/sk_curseval_tbl.py
  - openspec/changes/skcurseval/**
exclusive: []
gates:
  - "every claim tagged V read / V run / D / S / I / U; numbers reproducible with command, case count and seed"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "curse calc cells equal d2d on every level 1..99 (0 of 46,136 and 0 of 66,136 cases differ, seeds 1 and 7); no calc deviation, no PR candidate; ledger row 309-310; not verified live"
```

## 1. Work

- [x] 1.1 Oracle: game.exe FUN_00646ca0 on the 11 curse rows, levels 1..99, two seeds, vs d2d eval_calc
- [x] 1.2 Read the curse do for level and length divisor
- [x] 1.3 Trace doc (Q, D, B, limits, drill, matrix text), ledger line, coverage table regenerated
- [x] 1.4 Checks (changecheck, leakcheck)
