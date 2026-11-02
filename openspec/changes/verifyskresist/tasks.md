# Tasks

```yaml
change: verifyskresist
status: in-progress
agent: verifyskresist
branch: agent/verifyskresist
worktree: ~/code/d2lod-wt/verifyskresist
skip_specs: true
owns:
  - openspec/changes/verifyskresist/**
  - docs/research/verify-skresist.md
  - tools/emu/vr_dis.py
  - tools/emu/vr_tbl.py
  - tools/emu/vr_resist.py
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "Offline, not verified live. Own harness vr_resist.py: FUN_005c3540, FUN_0057be00, FUN_0057c6c0 run natively vs my own models, seeds 11 and 2025, 70,000 cases per part, 0 mismatches. Q4, Q5, Q6, Q8 HOLD; Q1, Q2, Q10 HOLD WITH LIMITS; Q3, Q7, Q9 not checked. d2d not run."
```

## 1. Work

- [x] 1.1 Own look at the four functions and own models
- [x] 1.2 Differential on two seeds
- [x] 1.3 Doc docs/research/verify-skresist.md
