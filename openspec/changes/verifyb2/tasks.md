# Tasks

```yaml
change: verifyb2
status: in-progress
agent: pa-verifyb2
branch: agent/verifyb2
worktree: d2lod-wt/verifyb2
skip_specs: true
owns:
  - docs/research/verify-batch2.md
  - tools/emu/vb2_*
  - openspec/changes/verifyb2/**
exclusive: []
gates:
  - "own harness: game.exe natively vs the real d2d builds base and head, counted"
  - "changecheck -cli 0 warnings; leakcheck -mode fail 0"
result: "Offline, nothing verified live. Minmana: own native harness, base 117 / 106 differ, head 0 (seeds 3, 11): ready with text changes. Teleport: tie order read in FUN_0064dea0 and run (0 differ on rings 1..49), but warp_spot searches ring 50 where game.exe stops at 49 (179 of 20,000): one-line fix before the PR. Suites, warnings, merge with #178: see docs/research/verify-batch2.md."
credit:
  component: VERIFYB2
  author: owner
  traced_by: []
  reported_by: []
  verified_by: []
  agents: [Claude Code]
  consent: none
```

## 1. Work

- [x] 1.1 Read FUN_0056c160, FUN_0056bfe0, FUN_005ca360, FUN_00554ea0, FUN_0064e7b0, FUN_0064dea0
- [x] 1.2 Own game-side harnesses (mana, warp spot search), own seeds
- [x] 1.3 Real d2d builds base vs heads, own probes, counts
- [x] 1.4 Commit hygiene, warnings, suites, merge with pull/178
- [x] 1.5 Compare with the author's counts; report docs/research/verify-batch2.md
