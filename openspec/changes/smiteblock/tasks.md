# Tasks

```yaml
change: smiteblock
status: in-progress
agent: smiteblock
branch: agent/smiteblock
worktree: ~/code/d2lod-wt/smiteblock
skip_specs: true
owns:
  - docs/research/smiteblock-trace.md
  - tools/emu/sb_block.py
  - tools/emu/sb_run.sh
  - tools/emu/sb_dbg.py
  - tools/emu/sb_d2d_test.cpp
  - tools/emu/sb_build.sh
  - tools/emu/sb_after.sh
  - openspec/changes/smiteblock/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs and emulator scripts only. A block does not stop Smite in game.exe 1.14d (V run, 96,000 cases, 23,347 blocked, all applied); d2d lets a block end it (V run, 34,916 of 60,000 blows with block > 0 differ, 0 after a one-line fix). No PR candidate (harness new). Nothing run live."
live: "not run live; emulator on the worker host only"
```

- [x] read the Smite / block path, FUN_0057ec10 flag mapping, gate and consumers
- [x] native run of the chain (seeds 7, 1, and a knockback variant) and d2d's player_blow before / after
- [x] trace doc, bot findings, drill recipe, matrix text, ledger note
