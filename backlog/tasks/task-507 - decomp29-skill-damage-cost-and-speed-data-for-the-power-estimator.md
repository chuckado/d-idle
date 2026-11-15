---
id: TASK-507
title: 'decomp29: skill damage, cost and speed data for the power estimator'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T15:45:00Z'
updated_date: '2026-11-15T16:33:00Z'
labels:
  - d2d
  - oracle
dependencies: []
priority: medium
ordinal: 277000
card: "decomp29"
owns:
  - "docs/research/decomp29-skilldamage.md"
  - "internal/power/skilldamage.json"
  - "tools/emu/dc29_"
exclusive: []
contracts: []
result: "131 class skills x levels 1..30 x 3 synergy sets: game.exe element/phys/calc/mana (V run) in internal/power/skilldamage.json; d2d element differential 24,750 cases 0 differ; rate tables FCR/IAS; no PR candidate; not wired in Go; not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/research/decomp29-skilldamage.md with V run tables
- [x] #2 internal/power/skilldamage.json schema documented
<!-- AC:END -->
