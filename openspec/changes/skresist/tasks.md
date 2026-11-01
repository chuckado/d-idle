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
result: ""
```

## 1. Work

- [ ] 1.1 Trace base resistances, curse / aura reduction, pierce, floor and damage in game.exe 1.14d (tags)
- [ ] 1.2 Read d2d's monster resist build, curse application and damage paths; the bot's immunity model
- [ ] 1.3 Oracle (game.exe natively) vs the real d2d build, two seeds
- [ ] 1.4 Trace doc with deviations, bot findings, drill recipe, matrix and shared-file lines
- [ ] 1.5 Checks (changecheck, leakcheck, openspec validate, ci check)
