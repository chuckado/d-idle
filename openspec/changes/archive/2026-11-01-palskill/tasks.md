# Tasks

```yaml
change: palskill
status: archived
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
result: "docs only, V run for flags, element damage, aura calcs and mana (0 differ on 8,640 and 21,760 cases); no PR candidate (no d2d-side test mode yet)"
merge: a439ea3
```

## 1. Work

- [x] 1.1 Read Holy Bolt hit function, Blessed Hammer do and damage chain, aura dos and pulse
- [x] 1.2 Differentials (flags, element damage, aura calcs, Concentration factor, hammer flight and hits)
- [x] 1.3 d2d source and bot comparison, deviations and findings
- [x] 1.4 Trace doc with drill recipe and matrix text
- [x] 1.5 merge-lines.txt, checks
