---
id: TASK-519
title: 'verifyp17: independent verifier for upstream candidate P17 (sequence -30 attack rate, fid/seq-ias-r1)'
status: Proposed
assignee:
  - verifyp17
created_date: '2026-11-16T01:31:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 282000
card: "verifyp17"
owns:
  - "docs/upstream/verifier-p17-2026-11-15.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P17 (fid/seq-ias-r1) must pass an independent verifier before it is posted upstream. Done when: the verdict doc names the exact head to post. Consumer: the upstream PR queue (P17 row).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 r2 rebased on current upstream main, rebuilt, oracle and claim grids rerun
- [ ] #2 game-side rule re-derived from game.exe and the grid axes judged
- [ ] #3 mutants, full ctest vs upstream main, tidy, presubmit, pr_guard, trailers checked
- [ ] #4 verdict and exact head in docs/upstream/verifier-p17-2026-11-15.md
<!-- AC:END -->
