# Tasks

```yaml
change: drills20
status: archived
agent: drills20
branch: agent/drills20
worktree: ~/code/d2lod-wt/drills20
skip_specs: true
owns:
  - openspec/changes/drills20/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
merge: 5516648
result: "step-away restart storm fixed live (walk timeouts 173 -> 0, hold runs, bounces stay 0); step-aways 129 -> 62 (target 20 not met): remaining roam leg retries are bounded (card ROAMLEG, recorded in BACKLOG)"
live: "DW.m3 x3 FAIL (restart storm unchanged), DW.m4 unchanged, D2 PASS, D3 PASS"
```

- [x] 1.1 DW.m3 x3 and DW.m4 live on the fixmouthloop build
- [x] 1.2 D2 and D3 regression
- [x] 1.3 root cause of the unchanged storm (walk.go:74 and :124)
