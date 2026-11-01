# Tasks

```yaml
change: drills21
status: in-progress
agent: drills21
branch: agent/drills21
worktree: ~/code/d2lod-wt/drills21
skip_specs: true
owns:
  - openspec/changes/drills21/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge:
result: "fixmouthloop2 PARTIAL live: 0 mouth bounces, 0 walk timeouts, back-off fires, step-aways 129 to 62 (target 20), 22 level-change errors from a pocket-stall escape; D3 12/12 and D2 11/11 PASS"
live: "DW.m3 x3 PARTIAL, DW.m4 unchanged, D2 PASS, D3 PASS (DRILLS21)"
```

- [x] 1.1 DW.m3 x3 and DW.m4 live on the fixmouthloop2 build
- [x] 1.2 D2 and D3 regression
- [x] 1.3 caller chain of the remaining step-aways (roam < explore < Explore)
