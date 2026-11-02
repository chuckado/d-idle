# Tasks

```yaml
change: covmissiles
status: in-progress
agent: covmissiles
branch: agent/covmissiles
worktree: ~/code/d2lod-wt/covmissiles
skip_specs: true
owns:
  - docs/research/covmissiles-trace.md
  - tools/emu/cm_*
  - openspec/changes/covmissiles/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; coverage.py --check; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs and emulator scripts only. V run on the worker host: hit roll of 8 more Sorceress skills 138,000 cases (d2d before 40,850 differ, after 0); every Missiles.txt row made by game.exe 13,680 cases (game vs table 0, d2d 6,040 compared 0 differ, 382 rows not loaded by d2d). Deviations D1-D4, bot findings B1-B3, no new PR candidate. Not verified live: nothing was run in a game."
live: "docs only, nothing run live"
```

- [x] read the coverage docs and the closest traces
- [x] hit-roll differential for the Sorceress skills not yet run (cm_run.sh hit)
- [x] table and flight differential for all 684 rows (cm_run.sh flight)
- [x] read d2d's loader and column use (cm_sweep.sh)
- [x] trace doc, ledger rows, regenerated coverage table
- [x] checks and merge lines
