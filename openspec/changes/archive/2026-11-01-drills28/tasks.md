# Tasks

```yaml
change: drills28
status: archived
merge: f85fcc3
agent: drills28
branch: agent/drills28
worktree: ~/code/d2lod-wt/drills28
skip_specs: true
owns:
  - openspec/changes/drills28/**
exclusive:
  - instance:drills28-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "step 0: d2d refreshes a set Frozen Armor state; armorcut code ran 300 s with the armor never absent; retry path not exercised (needs a melee-hold drill). Verified live (harness drills), retry not verified live."
live: "D28-0 PASS on both binaries; D26-A FAIL by its own spacing rule on both (harness artifact, no cut line); D2 11/11, D3 12/12"
```

- [x] 1.1 D28-0 idle refresh drill (group V)
- [x] 1.2 D26-A rerun on bin/d2d.lab-armorhit and bin/d2d.lab-nexthit2
- [x] 1.3 regression D2 11/11, D3 12/12
