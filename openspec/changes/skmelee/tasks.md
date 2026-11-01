# Tasks

```yaml
change: skmelee
status: in-progress
agent: skmelee
branch: agent/skmelee
worktree: ~/code/d2lod-wt/skmelee
skip_specs: true
owns:
  - docs/research/sk-melee-trace.md
  - tools/emu/sk_melee.py
  - openspec/changes/skmelee/**
exclusive: []
gates:
  - "emulator oracle: 0 differences from the reading on 100,000 cases a part, two seeds"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: ""
```

Overlaps (not edited here, lines for the merger in the trace's last section): docs/decomp-queue.md, docs/BACKLOG.md, docs/INDEX.md, CHANGELOG.md, docs/research/README.md, docs/OWNERS, docs/research/sorc-skill-matrix.md.

## 1. Work

- [ ] 1.1 Trace the plain Attack in game.exe 1.14d (to-hit, damage build, weapon share, Str bonus, shield, undead / demon terms)
- [ ] 1.2 Read d2d's `make_fighter` / `player_blow` / `hit_chance`; the bot's melee model
- [ ] 1.3 Emulator oracle vs the reading and vs d2d (100,000 cases, two seeds)
- [ ] 1.4 Trace doc: deviations, bot findings, drill recipe, matrix cells, shared-file lines

## 2. Open

- [ ] 2.1 Q3 (does the weapon's enhanced damage reach the player's stats 17 / 18?): drill R1
- [ ] 2.2 A d2d-side differential on a real build for D1 / D2 (not built)
