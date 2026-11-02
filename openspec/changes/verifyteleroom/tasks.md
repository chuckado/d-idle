# Tasks

```yaml
change: verifyteleroom
status: in-progress
agent: verifyteleroom
branch: agent/verifyteleroom
worktree: ~/code/d2lod-wt/verifyteleroom
skip_specs: true
owns:
  - docs/research/verifyteleroom-trace.md
  - tools/emu/vtr_room.py
  - tools/emu/vtr_dis.py
  - tools/emu/vtr_cmp.sh
  - openspec/changes/verifyteleroom/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Docs and emulator scripts only. Room rule HOLDS (150,225 cases, 0 violations); landing may leave the caster's zone; author's numbers reproduced. Run in the emulator on the worker host, nothing live."
live: "docs only, nothing run live"
```

- [x] own disassembly of FUN_005ca360, FUN_00554ea0, FUN_00463740, FUN_00619790, FUN_0064dea0
- [x] grid, irregular-room and chain runs; misc branch checks; author's numbers rerun
- [x] trace doc with verdicts, limits, drill recipe, matrix text
