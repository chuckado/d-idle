# Tasks

```yaml
change: treasure
status: archived
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
result: "docs and emulator tools only; see the doc for deviations"
live: "docs only, nothing run live"
merge: b8689ae
```

- [x] selection path read in game.exe (FUN_005a6600, FUN_0055afa0, FUN_00558200)
- [x] upstream drop differential re-run on new seeds; selection oracle written and run
- [x] exact Act 1 / 2 class odds (potions, scrolls, gold, items per kill) checked against native rolls
- [x] d2d deviation list, bot findings, drill recipe
- [x] ledger row, merge lines
