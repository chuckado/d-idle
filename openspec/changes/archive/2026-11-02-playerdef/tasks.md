# Tasks

```yaml
change: playerdef
status: archived
agent: playerdef
branch: agent/playerdef
worktree: ~/code/d2lod-wt/playerdef
skip_specs: true
owns:
  - docs/research/playerdef-trace.md
  - tools/emu/pd_dis.sh
  - tools/emu/pd_calc.py
  - openspec/changes/playerdef/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "docs and emulator tools only; credits runner; the d2d side is a port not the build"
live: "not a live game; emulator V run of FUN_00622720 / 006223f0 / 00622560 against a port of the d2d source (not the compiled build)"
merge: 9f9b32d
```

- [x] read upstream traces, d2d source, bot source
- [x] write docs/research/playerdef-trace.md
- [x] merge-lines.txt, OWNERS
- [x] differential of block / defence / attack rating (section 7b)
- [ ] differentials of the rest of section 6 (life, mana, resist clamp, DR order, block recovery)
