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
result: "offline only: combat sim tests pass; not verified live"
```

- [x] restart path found (engage approach -> move -> GoTo -> new walk -> step-away -> gate hold)
- [x] back-off in Fighter.move with one debug line per episode
- [x] sim tests (held mover)
- [ ] live proof (DW.m3)
