# Tasks

```yaml
change: verifyaidel
status: in-progress
agent: verifyaidel
branch: agent/verifyaidel
worktree: ~/code/d2lod-wt/verifyaidel
skip_specs: true
owns:
  - docs/research/verify-aidel.md
  - tools/emu/va_aidel.py
  - tools/emu/va_aidel_run.py
  - openspec/changes/verifyaidel/**
exclusive: []
gates:
  - changecheck and leakcheck clean
result: aidel table and FUN_005a73e0 re-verified under the emulator (352,800 cases, 0 differ); freeze-end path not covered; not verified live
```

- [x] 1 Table per difficulty
- [x] 2 Own harness for FUN_005a73e0
- [x] 3 Trace doc
- [ ] 4 FUN_005a7c20 / freeze-end path (open)
