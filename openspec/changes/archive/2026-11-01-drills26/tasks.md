# Tasks

```yaml
change: drills26
status: archived
merge: 6a2c69c
agent: drills26
branch: agent/drills26
worktree: ~/code/d2lod-wt/drills26
skip_specs: true
owns:
  - openspec/changes/drills26/**
exclusive:
  - instance:drills26-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "recast works (fixrecast/fixbuff live), but a recast cut at 232.4 s left the armor down about 3.5 s: ARMORCUT; verified live (DRILLS26), champion pack and boss check not run"
live: "group U D26-A FAIL (armor absent 3480 ms of 300 s, both binaries)"
```

- [x] 1.1 drill D26-A (group U)
- [x] 1.2 live run on bin/d2d.lab-armorhit and bin/d2d.lab-nexthit2
- [x] 1.3 boss-adjacent check: skipped, the harness has no Andariel stand drill
