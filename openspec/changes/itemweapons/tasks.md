# Tasks

```yaml
change: itemweapons
status: in-progress
agent: itemweapons
branch: agent/itemweapons
worktree: d2lod-wt/itemweapons
skip_specs: true
owns:
  - docs/research/itemweapons-trace.md
  - openspec/changes/itemweapons/**
exclusive: []
gates: trace doc exists, leakcheck and changecheck clean
result: "first pass: weapons.txt table facts from the MPQ (D); no emulator run, no ledger rows; not verified live"
credit:
  author: owner
  agents: [Claude Code]
```

- [x] 1. Dump weapons.txt and tabulate (D)
- [ ] 2. Run game.exe damage and speed formulas against d2d (not done)
- [ ] 3. itemtypes and armor tables
