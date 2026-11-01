# Tasks

```yaml
change: fixmouth2
status: in-progress
agent: fixmouth2
branch: agent/fixmouth2
worktree: ~/code/d2lod-wt/fixmouth2
skip_specs: true
owns:
  - internal/nav/walk.go
  - internal/nav/fixmouth_test.go
  - internal/combat/fight.go
  - internal/combat/fighter.go
  - internal/combat/fixmouth2_test.go
  - docs/fixmouth2.md
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4 rerun, no bounce, Clear finishes (docs/fixmouth2.md)"
merge: ""
result: "offline only: unit and sim tests pass; live proof pending"
```

- [x] decide the change from the DRILLS16 geometry
- [x] arrival held for the stay, NearMouth 12, evade cap
- [x] table tests and a Clear sim
- [ ] live proof (DW.m3 x3, DW.m4)
