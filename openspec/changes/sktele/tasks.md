# Tasks

```yaml
change: sktele
status: in-progress
agent: sktele
branch: agent/sktele
worktree: ~/code/d2lod-wt/sktele
skip_specs: true
owns:
  - docs/research/sk-teleport-trace.md
  - tools/emu/sk_tele_spot.py
  - openspec/changes/sktele/**
exclusive: []
gates:
  - "destination rule traced in game.exe (do 27, FUN_00554ea0) and run natively on synthetic rooms; every d2d difference classified"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "pending"
```

## 1. Work

- [ ] 1.1 Trace the start refusals, do 27 and the warp in game.exe 1.14d (tags)
- [ ] 1.2 Emulator oracle of the destination rule vs d2d's rule (line copy)
- [ ] 1.3 d2d's and the bot's Teleport, deviations and findings
- [ ] 1.4 Trace doc: drill recipe, matrix cells, shared-file lines for the merger

## 2. Open

- [ ] 2.1 Live drill (FORKDBG verbs) for the snap, the room limit and the free refused cast
- [ ] 2.2 A d2d-build harness (C++ test mode on a real Level) before any upstream candidate
