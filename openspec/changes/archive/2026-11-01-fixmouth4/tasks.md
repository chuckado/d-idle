# Tasks

```yaml
change: fixmouth4
status: archived
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
merge: 1787725
result: "den bounce closed live in DRILLS19 (DW.m3 0 bounces x3, DW.m4 0); earlier iterations improved it; trigger bound and long-click behaviour remain UNVERIFIED as principles"
live: "verified live (drills19, hidden slot 0, lab build cheats3): DW.m3 x3 0 bounces in 4 entries, 0 level-change errors, 351 rewrites (away) and 117 holds per run; DW.m4 x3 0 bounces, 1 rewrite (clip); D2 11/11, D3 12/12; trigger bound: no click fired from beyond 15 (no bounce to inspect): U; Clear and Explore finish; ErrMouthGate is returned (117 times) but never logged by a caller"
```

- [x] structural reason and design written first (docs/fixmouth4.md)
- [x] click gate on every nav world click (clip, away rewrite, bounded hold), named trigger bound with evidence
- [x] step-away replan re-check
- [x] swing gate in combat
- [x] tests: logged geometries incl. the DRILLS18 chain, property test, sims, wanted exits both ways
- [x] live proof (DW.m3 x3, DW.m4 x3): PASS, drills19
