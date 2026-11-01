# Tasks

```yaml
change: fixmouthloop2
status: in-progress
agent: fixmouthloop2
branch: agent/fixmouthloop2
worktree: ~/code/d2lod-wt/fixmouthloop2
skip_specs: true
owns:
  - internal/nav/fixmouthloop2_test.go
  - docs/fixmouthloop2.md
  - openspec/changes/fixmouthloop2/**
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4: 0 bounces, step-aways per run about 20 or fewer, no 2m0s walk timeouts, the backing off line appears (docs/fixmouthloop2.md)"
merge: ""
result: "offline: nav and combat tests pass, loop reproduced (682213 held clicks old, 1 new)"
live: "PARTIAL (DRILLS21, tip 293c816): V 0 mouth bounces, 0 walk timeouts (was 173), back-off line fires (1), Clear and Explore finish; NOT met: step-aways 62 per invocation (was 129, target about 20), 22 level-change errors from one pocket-stall town-portal escape; remaining storm is Navigator.roam (10 failed legs per episode)"
```

- [x] loop traced (walk.go step-away block, walkSeq vs mouthSeq)
- [x] fix: one attempt per walk, ErrMouthGate returned
- [x] sim tests
- [x] live proof (DW.m3 x3, DW.m4): PARTIAL, see DRILLS21
