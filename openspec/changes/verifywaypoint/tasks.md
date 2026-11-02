# Tasks

```yaml
change: verifywaypoint
status: in-progress
agent: verifywaypoint
branch: agent/verifywaypoint
worktree: ~/code/d2lod-wt/verifywaypoint
skip_specs: true
owns:
  - openspec/changes/verifywaypoint/**
  - docs/research/verify-waypoint.md
  - tools/emu/wpv_scan.py
  - tools/emu/wpv_ctx.py
  - tools/emu/wpv_dis.py
  - tools/emu/wpv_cmp.py
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
result: "Offline, not verified live. Writer/reader scan of game.exe (24 hits, one pcdata writer FUN_0055b720), 59,309 of 132,964 reproduced exactly (seed 7), second set seed 11 100000 cases: hostile rule 0 mismatches. New: FUN_00584870 reads the field with a 5 s window. Single player is inferred, not proven."
```

## 1. Work

- [x] 1.1 Scan and context of every +0x160 access
- [x] 1.2 Reproduce counts, second case set
- [x] 1.3 Write docs/research/verify-waypoint.md
