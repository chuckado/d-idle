# Tasks

```yaml
change: fixmouthloop
status: in-progress
agent: fixmouthloop
branch: agent/fixmouthloop
worktree: ~/code/d2lod-wt/fixmouthloop
skip_specs: true
owns:
  - internal/combat/mouthhold.go
  - internal/combat/fixmouthloop_test.go
  - docs/fixmouthloop.md
  - openspec/changes/fixmouthloop/**
exclusive: []
gates:
  - "live: DW.m3 with the gate: step-aways per Den cycle drop from 117 to a handful, the cycle does not burn the phase limit (docs/fixmouthloop.md)"
merge: ""
result: "offline only: combat sim tests pass; live proof FAILED (drills20)"
live: "NOT verified: DRILLS20 DW.m3 x3 counts equal DRILLS19 (129 step-aways, 0 back-off lines): the loop is inside one Navigator.walk (walk.go:74 and :124), the Fighter never sees ErrMouthGate; gate preserved (0 bounces, 0 level-change errors, D2 11/11, D3 12/12) DRILLS21 on fixmouthloop2 (which contains it): the Fighter back-off now runs live (1 backing off line, V)."
```

- [x] restart path found (engage approach -> move -> GoTo -> new walk -> step-away -> gate hold)
- [x] back-off in Fighter.move with one debug line per episode
- [x] sim tests (held mover)
- [ ] live proof (DW.m3)
