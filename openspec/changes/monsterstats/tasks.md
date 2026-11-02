# Tasks

```yaml
change: monsterstats
status: in-progress
agent: monsterstats
branch: agent/monsterstats
worktree: ~/code/d2lod-wt/monsterstats
skip_specs: true
owns:
  - docs/research/monster-stats-trace.md
  - tools/emu/ms_*
  - openspec/changes/monsterstats/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs and tools/emu/ms_* only. Offline emulator runs on the worker host (game.exe FUN_006538a0 / FUN_00573cb0 against d2d's real rules::monster_stats); see docs/research/monster-stats-trace.md. Not verified live."
live: "docs only, nothing run live"
```

- [x] game.exe stat rows (FUN_006538a0) and unit creation (FUN_00573cb0) run for all rows x difficulties x levels
- [x] d2d's real monster_stats compared (ms_d2d_test)
- [x] real population trace of Act 1 and 2 plain monsters (ms_spawn)
- [x] trace doc, bot findings, drill recipe, matrix lines
- [x] ledger row (case counts in the doc)
