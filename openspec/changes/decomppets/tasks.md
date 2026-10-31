# Tasks

```yaml
change: decomppets
status: in-progress
agent: decomppets
branch: agent/decomppets
worktree: ~/code/d2lod-wt/decomppets
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
result: ""
```

## 1. Work

- [ ] 1.1 Read the queue rules, the PR standard, the earlier traces, the #134 proof, the merc policy, primitives
- [ ] 1.2 game.exe: the monster target search, the merc think / attack / search, the pet thinks (V, with upstream's RE docs as S)
- [ ] 1.3 d2d: ai.cpp search_target, fight/merc.cpp, fight/pets.cpp, fight/tick.cpp foes (V-d2d)
- [ ] 1.4 Oracle: game.exe vs d2d's real rules code, >= 100,000 cases
- [ ] 1.5 Bot model: internal/ merc, combat, safety; SME docs; findings
- [ ] 1.6 Trace doc with DRILL RECIPE, decomp-queue row and result line, README / INDEX lines
- [ ] 1.7 changecheck 0 warnings, leakcheck fail 0, openspec validate --all --strict
