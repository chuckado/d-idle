# Tasks

```yaml
change: drills10
status: in-progress
agent: drills10
branch: agent/drills10
worktree: ~/code/d2lod-wt/drills10
skip_specs: true
owns:
  - openspec/changes/drills10/**
exclusive:
  - instance:slot0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; hygiene-only no FAIL; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: "F x3 PASS, D8.q D8.w PASS, D-meter PASS, P26 recorded, regression PASS"
```

- [x] 1.1 fixf17: full F group x3, DW.m1 run 1, Pacer lag lines
- [x] 1.2 fixtalk: D8 group (D8.q, D8.w)
- [x] 1.3 fixmeter: D-meter
- [x] 1.4 P26 observables (record only)
- [x] 1.5 regression pass (D2, D6, DW, H, F)
- [x] 1.6 docs, checks
