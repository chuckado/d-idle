# Tasks

```yaml
change: palskill
status: in-progress
agent: palskill
branch: agent/palskill
worktree: ~/code/d2lod-wt/palskill
skip_specs: true
owns:
  - docs/research/pal-skill-trace.md
  - tools/emu/pal_*
  - openspec/changes/palskill/**
exclusive: []
gates:
  - "game.exe natively vs the real d2d build: flags, element damage, aura calcs; every mismatch explained"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Offline, nothing verified live. Holy Bolt hits only undead and heals friendly units (V read), Blessed Hammer spirals out of the caster, gains half of Concentration's damage % and +50 % vs undead / demon (V read; factor and flight V run); undead / demon flags (734 ids), element damage (8,640) and aura calcs (21,760) equal d2d with 0 differ; deviations D1-D7 are rules d2d lacks (no differential today); bot B1-B5; no PR candidate."
```

## 1. Work

- [x] 1.1 Read Holy Bolt hit function, Blessed Hammer do and damage chain, aura dos and pulse
- [x] 1.2 Differentials (flags, element damage, aura calcs, Concentration factor, hammer flight and hits)
- [x] 1.3 d2d source and bot comparison, deviations and findings
- [x] 1.4 Trace doc with drill recipe and matrix text
- [x] 1.5 merge-lines.txt, checks
