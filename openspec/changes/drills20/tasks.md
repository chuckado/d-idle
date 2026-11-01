# Tasks

```yaml
change: drills20
status: in-progress
agent: drills20
branch: agent/drills20
worktree: ~/code/d2lod-wt/drills20
skip_specs: true
owns:
  - openspec/changes/drills20/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result: "fixmouthloop FAIL live: counts equal DRILLS19, 0 back-off lines; regression D3 12/12 and D2 11/11 PASS"
live: "DW.m3 x3 FAIL (restart storm unchanged), DW.m4 unchanged, D2 PASS, D3 PASS"
```

- [x] 1.1 DW.m3 x3 and DW.m4 live on the fixmouthloop build
- [x] 1.2 D2 and D3 regression
- [x] 1.3 root cause of the unchanged storm (walk.go:74 and :124)
