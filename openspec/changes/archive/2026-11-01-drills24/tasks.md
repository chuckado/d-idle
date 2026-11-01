# Tasks

```yaml
change: drills24
status: archived
merge: 194e1ab
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
result: "Teleport key wired from the bound skill (production never had it); hops tuned to pay (defaults 36 subtiles/40 percent reserve): live: near walking speed with the defaults, 18 ticks per hop; pockets/escapes and path-aware hops are the next card (TELETUNE2)"
live: "group T PASS, D2 11 of 11, D3 12 of 12 (DRILLS24)"
```

- [x] 1.1 d2drill knob flags, TeleportCost wiring, per-hop ticks, pocket drills
- [x] 1.2 live runs: default knobs, teleMinLeg 0 and 10, pocket legs
- [x] 1.3 regression D2, D3
