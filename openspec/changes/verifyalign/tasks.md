# Tasks

```yaml
change: verifyalign
status: in-progress
agent: verifyalign
branch: agent/verifyalign
worktree: d2lod-wt/verifyalign
skip_specs: true
owns:
  - openspec/changes/verifyalign/**
  - docs/research/verify-alignment2.md
  - tools/emu/vfal_run.py
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "Offline, not verified live. Own harness vfal_run.py (seeds 20261102 and 77): getter/setter 400 + 400 random units, gate FUN_0057b7d0 600 + 600 runs, spawn map 734 MonStats rows, 21 setter call sites: 0 differ; no claim of alignment2-trace.md found WRONG; open: live stat 0xac on merc/Shadow, merc weapon path of FUN_00535bc0, two unattributed value-2 setter sites."
credit:
  component: verifyalign
  author: owner
  verified_by: []
  agents: [Claude Code]
```

## 1. Work

- [x] 1.1 Own disassembly reading of the gate, getter, setter, spawn map
- [x] 1.2 Own emulator harness, two seeds
- [x] 1.3 Verdict doc, card, merge lines, OWNERS
