# Tasks

```yaml
change: sktele
status: archived
merge: 6a6eabe
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
result: "Offline, nothing verified live. Teleport traced in game.exe (start FUN_0056f640, do 27 FUN_005ca360, warp FUN_00554ea0) and the destination rule run natively on synthetic rooms (tools/emu/sk_tele_spot.py): range = the caster's room or its near list, landing snapped to the nearest free spot (mask 0x1c09 incl. units), refused hops free, Levels Teleport 2 = clear line. Against a line copy of d2d's spot(): 31,248 of 100,000 differ, all D1 room limit (25,120), D3 units (4,805), D2 snap (1,323); D4-D7 read. No PR candidate (no real-build harness). Bot B1-B7 (B1: 18 subtiles is the screen clamp; B6: lair read as forbidden). Shared-file lines for the merger at the end of the trace."
```

## 1. Work

- [x] 1.1 Trace the start refusals, do 27 and the warp in game.exe 1.14d (tags)
- [x] 1.2 Emulator oracle of the destination rule vs d2d's rule (line copy)
- [x] 1.3 d2d's and the bot's Teleport, deviations and findings
- [x] 1.4 Trace doc: drill recipe, matrix cells, shared-file lines for the merger

## 2. Open

- [ ] 2.1 Live drill (FORKDBG verbs) for the snap, the room limit and the free refused cast
- [ ] 2.2 A d2d-build harness (C++ test mode on a real Level) before any upstream candidate
