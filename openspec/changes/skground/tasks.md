# Tasks

```yaml
change: skground
status: in-progress
agent: skground
branch: agent/skground
worktree: ~/code/d2lod-wt/skground
skip_specs: true
owns:
  - docs/research/sk-ground-trace.md
  - docs/research/skground-trace.md
  - tools/emu/sk_ground*
  - tools/emu/skg_*
  - openspec/changes/skground/**
exclusive: []
gates:
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0; tools/ci/check.sh; openspec validate --all --strict"
result: "Docs and offline tooling: docs/research/skground-trace.md (third pass, Meteor's blast shape: Q26-Q32, M1-M4, B1-B3, drill G1-G4); game.exe's FUN_0056bad0 / FUN_0056b7e0 run natively (disc on integer subtiles 0 differ in 120,000 cases, radius 6 at 20 of 20 levels, caster-room rule 0 of 168 far, filter modes) and compared with the real d2d Fight::area (9.1 percent of 80,000 rim cases differ, no bias); no PR candidate. Not verified live."
live: "docs only, nothing run live"
```

- [x] Meteor blast read (FUN_005aabb0, FUN_0056bad0, FUN_0056b7e0, FUN_0056b3e0, FUN_006492a0)
- [x] game-side native run (shape, rooms, filter, radius) and real-d2d differential (120,000 cases)
- [x] deviations M1-M4, bot findings, drill recipe, matrix cells
- [x] coverage ledger rows for Meteor (56) and meteorcenter (101); merge-lines.txt
