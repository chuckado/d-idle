# Tasks

```yaml
change: drills24
status: in-progress
agent: drills24
branch: agent/drills24
worktree: ~/code/d2lod-wt/drills24
skip_specs: true
owns:
  - openspec/changes/drills24/**
exclusive:
  - instance:drills24-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "teletune proven live: default knobs 0 hops (265 and 605 ticks against walking 254 and 578), teleMinLeg 0 stops at the mana reserve, 18 ticks per hop; pocket leg shows what Teleport is for and that the leg rule blocks it"
live: "group T PASS, D2 11 of 11, D3 12 of 12 (DRILLS24)"
```

- [x] 1.1 d2drill knob flags, TeleportCost wiring, per-hop ticks, pocket drills
- [x] 1.2 live runs: default knobs, teleMinLeg 0 and 10, pocket legs
- [x] 1.3 regression D2, D3
