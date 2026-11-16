---
id: TASK-528
title: 'verifyp18: independent verifier of P18 Smite range (fid/smite-r2)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-16T02:30:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: high
ordinal: 283000
card: "verifyp18"
owns:
  - "docs/upstream/verifier-p18-2026-11-15.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Consumer: upstream-pr P18 (smite-range). Opus verifier before posting P18 (Smite: stat 111 and max >= min + 1 point). Done when the verifier doc gives a verdict and the head to post.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 fid/smite-r2 rebased on upstream main, rebuilt, grid compare on oracle and claim, mutants, full ctest vs upstream main
- [ ] #2 Smite range rule re-derived from game.exe (FUN_0057b420, FUN_005ce9f0), grid axes and test_combat bounds judged
- [ ] #3 Body honesty checked; verdict and exact head in docs/upstream/verifier-p18-2026-11-15.md
<!-- AC:END -->
