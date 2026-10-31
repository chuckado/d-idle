# Tasks

```yaml
change: drills7
status: in-progress
agent: drills7
branch: agent/drills7
worktree: ~/code/d2lod-wt/drills7
skip_specs: true
owns:
  - openspec/changes/drills7/**
  - docs/probes/drills/*drills7*
  - cmd/d2drill/**
  - internal/drill/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: ""
```

- [ ] 1.1 H.pause x3 invocations (6 tries)
- [ ] 1.2 D6.d x3
- [ ] 1.3 D-sync arrival variants (town portal, waypoint)
- [ ] 1.4 full D2 group in one process (harness fix), twice
- [ ] 1.5 P26 observables (record only)
