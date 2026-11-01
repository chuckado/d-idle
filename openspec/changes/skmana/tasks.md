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
  - docs/decomp-queue.md
  - docs/BACKLOG.md
  - docs/INDEX.md
  - openspec/changes/skmana/**
exclusive: []
gates:
  - "pure differential game.exe vs the real d2d build: every Skills.txt row x levels 1..60 x mana thresholds plus random cases; every mismatch explained"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: ""
```

## 1. Work

- [ ] 1.1 Trace the check, the payment and their timing in game.exe 1.14d (tags)
- [ ] 1.2 Regen, potion bonus, skill delay, leech pointers
- [ ] 1.3 Pure differential game.exe vs the real d2d build
- [ ] 1.4 d2d's and the bot's mana model, deviations and findings
- [ ] 1.5 Docs: trace (drill recipe, matrix cells), queue row 19, BACKLOG, INDEX

## 2. Open

- [ ] 2.1 Live drill (FORKDBG verbs) for the pay timing and the free cast
