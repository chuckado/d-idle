# Tasks

```yaml
change: verifypalaura
status: in-progress
agent: verifypalaura
branch: agent/verifypalaura
worktree: d2lod-wt/verifypalaura
owns:
  - docs/research/verifypalaura.md
  - tools/emu/vp_check.py
  - openspec/changes/verifypalaura/**
skip_specs: true
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
merge:
result: "Emulator runs on the worker host (roll 4000 pulses, disc radius 7 and 16); no live game."
```

## 1. Work

- [x] 1.1 own roll and disc runs
- [ ] 1.2 filter, Prayer, regen, resist re-runs (next pass)
