# Tasks

```yaml
change: seeddist
status: in-progress
agent: seeddist
branch: agent/seeddist
worktree: ~/code/d2lod-wt/seeddist
skip_specs: true
owns:
  - openspec/changes/seeddist/**
  - docs/research/seeddist-trace.md
  - tools/emu/sd_*
exclusive: []
gates:
  - "changecheck 0 warnings for this card; leakcheck -mode fail 0"
  - "every V run number has stated cases, seeds and the test used"
result: "Offline, nothing verified live, no d2d slot, no game window, no Go change. Doc docs/research/seeddist-trace.md: for 8 Act 1 classes (Fallen, Zombie, Skeleton, Quill Rat, Rogue/Dark Hunter, Gargantuan Beast, Blood Raven, Andariel) game.exe 1.14d's real functions under the emulator vs d2d's real compiled code (scratch upstream build): life roll 100,000 units per class per side, hit roll 4 defence settings x 20,000+ swings, damage roll ~20,900 hits, first drop 3 configurations x 20,000 per class (540,000 per side), seed-for-seed control 360,000 drop and 160,000 life cases (0 differ in value), real room population cross-check (83,000+ units). 330 p-values, 16 below 0.05 (16.5 expected), none below 0.01: distributions equal; deviations D1 (life on the world stream), D2 (fixed stream start), D3 (damage roll shape, variance ratio 1.2 to 3.0, means equal), D4 (fixed-life rows step the seed in game.exe only); bot findings B1..B4; no PR candidate; ledger row formulas F16 TESTED-0-DIFF."
```

## 1. Work

- [x] 1.1 Read seed-study, d2d life / drop seeding, game.exe unit seeding (0x552df0, 0x45c3e0, 0x573f83)
- [x] 1.2 Game-side sampler (sd_game.py) and d2d-side sampler (sd_d2d_test.cpp, sd_build.sh)
- [x] 1.3 Tests (chi-square, KS, z) per quantity, 8 classes x >= 20,000 (sd_stats.py)
- [x] 1.4 Trace doc, ledger row F16 and regenerated rule-coverage.md, merge-lines.txt
