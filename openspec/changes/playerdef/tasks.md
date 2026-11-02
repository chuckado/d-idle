# Tasks

```yaml
change: playerdef
status: in-progress
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
result: "Trace plus upgrade run: block, defence, attack rating of game.exe run natively (pd_calc.py, 102,604 cases: AR 0 differ, block 0 differ after the floor at 0, defence 0 differ for positive base; D9 negative-base defence); D1..D8 still S/I/U; 4 bot findings; no PR candidate; one ledger row."
live: "not a live game; emulator V run of FUN_00622720 / 006223f0 / 00622560 against a port of the d2d source (not the compiled build)"
```

- [x] read upstream traces, d2d source, bot source
- [x] write docs/research/playerdef-trace.md
- [x] merge-lines.txt, OWNERS
- [x] differential of block / defence / attack rating (section 7b)
- [ ] differentials of the rest of section 6 (life, mana, resist clamp, DR order, block recovery)
