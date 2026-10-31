# Tasks

```yaml
change: decomppets
status: archived
agent: decomppets
branch: agent/decomppets
worktree: ~/code/d2lod-wt/decomppets
merge: agent/decomppets (merge commit this branch tip)
skip_specs: true
owns:
  - docs/research/pets-merc-trace.md
  - docs/research/pr-pets-*.md
  - tools/emu/pets_merc*
  - openspec/changes/decomppets/**
  - docs/decomp-queue.md
  - docs/research/README.md
  - docs/INDEX.md
  - docs/research/pr-bundle-queue.md
exclusive: []
gates:
  - "every rule in the trace doc is tagged V, D, S, I or U with an address or file:line"
  - "a pure rule is run in game.exe's own code against d2d's real code on >= 100,000 cases; a stateful AI gets the best verified structural comparison and a drill recipe"
  - "go run ./cmd/changecheck -cli has 0 warnings; go run ./cmd/leakcheck -mode fail finds 0; tools/openspec.sh validate --all --strict passes"
result: "Monsters (evil) take the strictly nearest of the player then its pets newest first, ties to the player, the player at >= 0x37 / town / another act drops its pets too, no threat score; the merc leashes (100 teleport / 24 run / 16 follow) before it fights a foe under 25, the Rogue backs off under 4, no low-life retreat, reacts at its next think (<= 10 frames). Differential vs d2d's real rules C++ (pets_merc.py + pets_merc_harness.cpp): merc target 0 of 100,000; monster search 1,254 of 100,000 differ (all pet ties, d2d merc first then oldest) -> 0 newest first, rule level, PR not built; upstream oracles re-run at 100k each ok. Bot findings B1-B5 (the Rogue is no tank; the merc workaround breaks on the leash); drill recipe T1-T6. Not verified live; nothing pushed or posted."
```

## 1. Work

- [x] 1.1 Read the queue rules, the PR standard, the earlier traces, the #134 proof, the merc policy, primitives
- [x] 1.2 game.exe: the monster target search, the merc think / attack / search, the pet thinks (V, with upstream's RE docs as S)
- [x] 1.3 d2d: ai.cpp search_target, fight/merc.cpp, fight/pets.cpp, fight/tick.cpp foes (V-d2d)
- [x] 1.4 Oracle: game.exe vs d2d's real rules code, >= 100,000 cases
- [x] 1.5 Bot model: internal/ merc, combat, safety; SME docs; findings
- [x] 1.6 Trace doc with DRILL RECIPE, decomp-queue row and result line, README / INDEX lines
- [x] 1.7 changecheck 0 warnings, leakcheck fail 0, openspec validate --all --strict
