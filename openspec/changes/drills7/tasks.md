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
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; whole-repo vet and test -short on the worker host"
merge:
result: ""
live: "H.pause FAIL (3 of 6), D6.d FAIL x3, D-sync.tp PASS, D-sync.wp FAIL 1 of 2, full D2 group PASS x2 after harness fix, P26 probe partial; findings F10 to F12"
```

- [x] 1.1 H.pause x3 invocations (6 tries)
- [x] 1.2 D6.d x3
- [x] 1.3 D-sync arrival variants (town portal, waypoint)
- [x] 1.4 full D2 group in one process (harness fix), twice
- [x] 1.5 P26 observables (record only, partial)
