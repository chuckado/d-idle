# Tasks

```yaml
change: verifyfreezeend2
status: in-progress
agent: verifyfreezeend2
branch: agent/verifyfreezeend2
worktree: ~/code/d2lod-wt/verifyfreezeend2
skip_specs: true
owns:
  - openspec/changes/verifyfreezeend2/**
  - docs/research/verifyfreezeend2-trace.md
  - tools/emu/vfe2_run.py
  - tools/emu/vfe2_cmp.py
  - tools/emu/vfe2_tab.py
  - tools/emu/vfe2_all.sh
  - tools/emu/vfe2_build.sh
  - tools/emu/test_vfe2_gh.cpp
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "Offline, not verified live. game.exe run natively under the emulator: think due during a get-hit 3,000 cases (1,271 inside) driver entered at the due frame, 0 differ; freeze during get-hit 3,000 cases (1,952 end event dropped) 0 differ from the stated rule; d2d probe differential 2,504 cases: think inside the get-hit 1,133 of 1,133 differ (d2d holds it to the get-hit end), after it 0 of 1,371. No PR candidate. Coverage ledger unchanged."
```

## 1. Work

- [x] 1.1 Read the first pass docs, trace the get-hit start and the driver
- [x] 1.2 Own harness, game side (gh, frz) run on the worker host
- [x] 1.3 d2d probe built and run, compared per case
- [x] 1.4 Trace doc, card, OWNERS, merge lines
