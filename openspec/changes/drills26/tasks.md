# Tasks

```yaml
change: drills26
status: in-progress
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
result: "D26-A live on lab-armorhit and lab-nexthit2: identical bot behaviour (plain zombies are frozen on both); one recast inside the second expiry did not refresh the armor (3.5 s absent); no boss drill exists, skipped"
live: "group U D26-A FAIL (armor absent 3480 ms of 300 s, both binaries)"
```

- [x] 1.1 drill D26-A (group U)
- [x] 1.2 live run on bin/d2d.lab-armorhit and bin/d2d.lab-nexthit2
- [x] 1.3 boss-adjacent check: skipped, the harness has no Andariel stand drill
