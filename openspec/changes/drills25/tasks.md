# Tasks

```yaml
change: drills25
status: in-progress
agent: drills25
branch: agent/drills25
worktree: ~/code/d2lod-wt/drills25
skip_specs: true
owns:
  - openspec/changes/drills25/**
exclusive:
  - instance:drills25-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "teletune2 live: Stony Tomb 211 and Halls 533 ticks beat walking (254, 578); border 717 vs 708 (+9, 0 hops); pocket leaves by hop in 233 ticks; stall escape not exercised live"
live: "group T default knobs, pocket, D2 11 of 11, D3 12 of 12 (DRILLS25)"
```

- [x] 1.1 live runs: group T default knobs, pocket start, controls
- [x] 1.2 regression D2, D3
- [x] 1.3 logs, README section, card lines
