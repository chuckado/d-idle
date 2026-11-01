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
  - docs/probes/drills/*drills10*
  - docs/probes/drills/README.md
  - docs/primitives.md
  - cmd/d2drill/**
  - internal/drill/**
exclusive:
  - instance:slot0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; hygiene-only no FAIL; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: ""
```

- [ ] 1.1 fixf17: full F group x3, DW.m1 run 1, Pacer lag lines
- [ ] 1.2 fixtalk: D8 group (D8.q, D8.w)
- [ ] 1.3 fixmeter: D-meter
- [ ] 1.4 P26 observables (record only)
- [ ] 1.5 regression pass (D2, D6, DW, H, F)
- [ ] 1.6 docs, checks
