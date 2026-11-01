# Tasks

```yaml
change: drills29
status: archived
agent: drills29
branch: agent/drills29
worktree: ~/code/d2lod-wt/drills29
skip_specs: true
owns:
  - openspec/changes/drills29/**
exclusive:
  - instance:drills29-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test on the worker host"
merge: 54b1d9c
result: "cut-cast retry proven live (DRILLS29: cut at 114.0 s, retry 1.6 s later, state present; armor absent 0 ms over 544 samples); a cut by real monsters at normal damage is unproven (needed a logged debug damage boost)"
live: "D29-C PASS on lab-armorhit and lab-nexthit2 (1 cut, 1 retry, armor never absent); D26-A PASS (spacing met) on lab-armorhit; D2 11/11, D3 12/12"
```

- [x] 1.1 D29-C drill and offline verdict test
- [x] 1.2 D26-A dead zone fix
- [x] 1.3 Live runs (D29-C, D26-A, nexthit2 comparison, D2 and D3 regression)
- [x] 1.4 Docs: README DRILLS29, primitives row, armorcut live line
