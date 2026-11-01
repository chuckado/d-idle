# Tasks

```yaml
change: drills22
status: archived
merge: 194e1ab
agent: drills22
branch: agent/drills22
worktree: ~/code/d2lod-wt/drills22
skip_specs: true
owns:
  - openspec/changes/drills22/**
exclusive:
  - instance:drills-slot-0
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "Teleport key wired from the bound skill (production never had it); hops tuned to pay (defaults 36 subtiles/40 percent reserve): live: near walking speed with the defaults, 18 ticks per hop; pockets/escapes and path-aware hops are the next card (TELETUNE2)"
live: "T group PASS 5 of 5, D2 11 of 11, D3 12 of 12, DW.m3 as baseline (1 bounce, not Teleport related) (DRILLS22)"
```

- [x] 1.1 group T drills (control, Teleport legs, bind, town, refused hops)
- [x] 1.2 live runs on the fixtelekey build
- [x] 1.3 regression D2, D3, DW.m3
