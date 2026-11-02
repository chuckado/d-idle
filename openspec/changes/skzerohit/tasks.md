# Tasks

```yaml
change: skzerohit
status: in-progress
agent: skzerohit
branch: agent/skzerohit
worktree: ~/code/d2lod-wt/skzerohit
skip_specs: true
owns:
  - docs/research/sk-zero-hit-trace.md
  - tools/emu/skzh_table.py
  - openspec/changes/skzerohit/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Offline, nothing verified live, nothing V run. Static answer (V read) in docs/research/sk-zero-hit-trace.md: FUN_0057cee0 runs for a zero-damage hit, knockback ignores damage, get-hit needs one point, no wake, Lightning Enchanted on-hit bolts fire only on a no-recovery hit; D1 to D3 and a drill recipe listed, no PR candidate."
```

## 1. Work

- [x] 1.1 Read FUN_0057cee0, FUN_005ad730, FUN_0057cb00, FUN_005a43a0 and the enchant event table
- [x] 1.2 Compare with d2d knockback and enchant code (read-only grep)
- [x] 1.3 Write the trace with tags, deviations, recipe
- [ ] 1.4 Emulator differential for D1 / D2 (open, section 6 of the trace)
