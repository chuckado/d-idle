# Tasks

```yaml
change: fixmouth3
status: in-progress
agent: fixmouth3
branch: agent/fixmouth3
worktree: ~/code/d2lod-wt/fixmouth3
skip_specs: true
owns:
  - internal/nav/roam.go
  - internal/nav/fixmouth3_test.go
  - internal/combat/fixmouth3_test.go
  - docs/fixmouth3.md
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4 rerun, zero level-change errors in Clear and Explore (docs/fixmouth3.md)"
merge: ""
result: "offline only: unit and sim tests pass; live proof pending"
```

- [x] root cause from the DRILLS17 logs (docs/fixmouth3.md)
- [x] stay guard, trigger radius, away-path check, one-click step-away, roam target filter
- [x] table tests and an Explore sweep sim
- [ ] live proof (DW.m3 x3, DW.m4)
