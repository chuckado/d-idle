# Tasks

```yaml
change: skmana
status: in-progress
agent: skmana
branch: agent/skmana
worktree: ~/code/d2lod-wt/skmana
skip_specs: true
owns:
  - docs/research/sk-mana-trace.md
  - tools/emu/sk_mana.py
  - tools/emu/sk_mana_test.cpp
  - openspec/changes/skmana/**
exclusive: []
gates:
  - "pure differential game.exe vs the real d2d build: every Skills.txt row x levels 1..60 x mana thresholds plus random cases; every mismatch explained"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "Offline, nothing verified live. Check / payment / timing / regen / potions / delay traced (docs/research/sk-mana-trace.md): cost formula equal to d2d on 130,840 cases (game.exe natively vs the real mana_cost, upstream 72132b99); D2 the check has no minmana floor and a short payment is skipped (free cast), 791 cases all explained, no plan impact; D1 / D3 paid at the action frame and free when interrupted (read, drills M1 M2); bot B1-B6 (B2 regen model 1.2-1.8x fast early); no PR candidate."
```

## 1. Work

- [x] 1.1 Trace the check, the payment and their timing in game.exe 1.14d (tags)
- [x] 1.2 Regen, potion bonus, skill delay, leech pointers
- [x] 1.3 Pure differential game.exe vs the real d2d build
- [x] 1.4 d2d's and the bot's mana model, deviations and findings
- [x] 1.5 Docs: trace (drill recipe, matrix cells), queue row 19, BACKLOG, INDEX

## 2. Open

- [ ] 2.1 Live drill (FORKDBG verbs) for the pay timing and the free cast
