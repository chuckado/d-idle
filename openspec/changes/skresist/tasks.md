# Tasks

```yaml
change: skresist
status: in-progress
agent: skresist
branch: agent/skresist
worktree: ~/code/d2lod-wt/skresist
skip_specs: true
owns:
  - docs/research/sk-resist-trace.md
  - tools/emu/sk_resist.py
  - tools/emu/sk_resist_test.cpp
  - openspec/changes/skresist/**
exclusive: []
gates:
  - "every claim tagged V read / V run / D / S / I / U; numbers reproducible (>= 100,000 cases or exhaustive boundaries, 32-bit masked seeds)"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict; tools/ci/check.sh"
result: "Offline, nothing verified live. Monster resist build traced (docs/research/sk-resist-trace.md): base = MonStats column of the difficulty, no penalty / cap for monsters, immune at total >= 100, pierce only under 100, floor -100; curses / auras on an immune / 5 (FUN_005c3540). D1 d2d applies the full value: real build upstream 72132b99, 147,628 cases, 88,101 adj / 14,355 dealt differ (seed 7: 88,224 / 14,289), 0 outside the rule; D2 (unreachable), D3 melee (read), D4 Sanctuary; bot B1-B5; no PR built."
```

## 1. Work

- [x] 1.1 Trace base resistances, curse / aura reduction, pierce, floor and damage in game.exe 1.14d (tags)
- [x] 1.2 Read d2d's monster resist build, curse application and damage paths; the bot's immunity model
- [x] 1.3 Oracle (game.exe natively) vs the real d2d build, two seeds
- [x] 1.4 Trace doc with deviations, bot findings, drill recipe, matrix and shared-file lines
- [x] 1.5 Checks (changecheck, leakcheck, openspec validate, ci check)
