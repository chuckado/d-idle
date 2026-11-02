# Tasks

```yaml
change: verifywaypoint
status: archived
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
result: "docs only; see the doc for HOLDS LIMITS FAILS"
merge: 35ca045
```

## 1. Work

- [x] 1.1 Scan and context of every +0x160 access
- [x] 1.2 Reproduce counts, second case set
- [x] 1.3 Write docs/research/verify-waypoint.md
