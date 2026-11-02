# Tasks

```yaml
change: palaura
status: in-progress
agent: palaura
branch: agent/palaura
worktree: ~/code/d2lod-wt/palaura
skip_specs: true
owns:
  - docs/research/palaura-trace.md
  - tools/emu/palaura_*
  - openspec/changes/palaura/**
exclusive: []
gates:
  - "game.exe's own aura code run under the emulator on real player and monster units (area search, filter, apply, enemy hit, Prayer mana and state 85, regeneration), every claim tagged"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; openspec validate --all --strict"
result: "Offline, nothing verified live. Game side V run only (no d2d differential): area search = integer subtile disc d2 <= r2 (2,028 of 2,032 equal, 4 outside near rooms); filter bits 0x12003 / 0xa783 / 0x4a283 decoded (MonStats bits boss, primeevil, npc, noAura by data); enemy hit one roll min..max-1 (6,000 pulses) and roll x (100 - (resist - pierce)) / 100; Holy Freeze skips prime evil; Prayer pays mana and sets state 85 only when a life rose, no heal below the cost, state 85 zeroes base and percent regen not the flat bonus. Deviations D1..D9, bot B1..B3, no PR candidate, drills P1..P4."
```

## 1. Work

- [x] 1.1 Read FUN_0056b7e0, FUN_0056b3e0, FUN_005cedc0, FUN_005cf010, FUN_005cf3a0, FUN_005806f0 and helpers
- [x] 1.2 Emulator runs: area radius, filter, apply, enemy hit, Prayer cases, regeneration, town, line of sight
- [x] 1.3 d2d's version (source), deviations, bot findings
- [x] 1.4 Trace doc, drill recipe, matrix text, merge lines
