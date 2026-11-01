# Tasks

```yaml
change: fixmouth4
status: in-progress
agent: fixmouth4
branch: agent/fixmouth4
worktree: ~/code/d2lod-wt/fixmouth4
skip_specs: true
owns:
  - internal/nav/mouthgate.go
  - internal/nav/fixmouth4_test.go
  - internal/combat/fixmouth4_test.go
  - docs/fixmouth4.md
  - openspec/changes/fixmouth4/**
exclusive: []
gates:
  - "live: DW.m3 x3 and DW.m4 rerun, zero mouth-bounce level-change errors and zero Entered cycles within 1 s (docs/fixmouth4.md)"
merge: ""
result: "offline only: unit, property and sim tests pass (internal/nav, internal/combat); not verified live"
```

- [x] structural reason and design written first (docs/fixmouth4.md)
- [x] click gate on every nav world click (clip, away rewrite, bounded hold), named trigger bound with evidence
- [x] step-away replan re-check
- [x] swing gate in combat
- [x] tests: logged geometries incl. the DRILLS18 chain, property test, sims, wanted exits both ways
- [ ] live proof (DW.m3 x3, DW.m4)
