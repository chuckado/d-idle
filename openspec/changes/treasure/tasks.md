# Tasks

```yaml
change: treasure
status: in-progress
agent: treasure
branch: agent/treasure
worktree: ~/code/d2lod-wt/treasure
skip_specs: true
owns:
  - docs/research/treasure-trace.md
  - tools/emu/tre_*
  - openspec/changes/treasure/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs and offline tooling: docs/research/treasure-trace.md (Q1-Q11, D1-D8, B1-B5, drill T1-T5). Run on the worker host: the upstream drop differential against the pristine d2d build (100,000 rolls over new seeds + 1,012 class tables + 20,000 objects + 20,000 stands + 39,980 items: 0 differ except 5 elixirs), a new selection oracle tools/emu/tre_select.py (78,309 selection cases and 85,200 upgrade cases, 0 differ from a transcription of d2d's rule), exact class odds tools/emu/tre_act.py checked by 104,000 native rolls. No d2d deviation proven, no PR candidate. Not verified live."
live: "docs only, nothing run live"
```

- [x] selection path read in game.exe (FUN_005a6600, FUN_0055afa0, FUN_00558200)
- [x] upstream drop differential re-run on new seeds; selection oracle written and run
- [x] exact Act 1 / 2 class odds (potions, scrolls, gold, items per kill) checked against native rolls
- [x] d2d deviation list, bot findings, drill recipe
- [x] ledger row, merge lines
