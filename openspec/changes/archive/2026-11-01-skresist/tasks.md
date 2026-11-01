# Tasks

```yaml
change: skresist
status: archived
merge: 44ff859
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
result: "resistance and immunity rules traced and oracle-verified; d2d curse-on-immune deviation proven on a real build (88,101 of 147,628): a local PR candidate card SKRESISTFIX is the next step; not verified live"
```

## 1. Work

- [x] 1.1 Trace base resistances, curse / aura reduction, pierce, floor and damage in game.exe 1.14d (tags)
- [x] 1.2 Read d2d's monster resist build, curse application and damage paths; the bot's immunity model
- [x] 1.3 Oracle (game.exe natively) vs the real d2d build, two seeds
- [x] 1.4 Trace doc with deviations, bot findings, drill recipe, matrix and shared-file lines
- [x] 1.5 Checks (changecheck, leakcheck, openspec validate, ci check)
