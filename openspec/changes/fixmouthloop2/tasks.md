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
live: "NOT verified: no game in this run; recipe in docs/fixmouthloop2.md"
```

- [x] loop traced (walk.go step-away block, walkSeq vs mouthSeq)
- [x] fix: one attempt per walk, ErrMouthGate returned
- [x] sim tests
- [ ] live proof (DW.m3 x3, DW.m4)
