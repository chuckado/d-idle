# Tasks

```yaml
change: drills8
status: in-progress
agent: drills8
branch: agent/drills8
worktree: ~/code/d2lod-wt/drills8
skip_specs: true
owns:
  - openspec/changes/drills8/**
exclusive:
  - instance:i0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: ""
```

- [ ] 1.1 H.pause x3 invocations
- [ ] 1.2 D6.d x3
- [ ] 1.3 D-sync waypoint arrival x3
- [ ] 1.4 NPC interaction smoke (D3, D8)
- [ ] 1.5 full drill suite regression table
