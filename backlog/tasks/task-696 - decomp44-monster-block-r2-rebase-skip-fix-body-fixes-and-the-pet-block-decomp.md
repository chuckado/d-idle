---
id: TASK-696
title: 'decomp44: monster block r2 rebase, skip fix, body fixes and the pet block decomp'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T18:47:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: high
ordinal: 337000
card: "decomp44"
owns:
  - "docs/research/decomp44-pet-block-trace.md"
  - "docs/research/decomp43-monster-block-trace.md"
  - "docs/upstream/bodies/monster-block-body.md"
exclusive: []
contracts: []
result: "fid/monster-block-r2 (local + private fork remote, on upstream c1b38311; r1 untouched): rebased with the gear_res conflict resolved, test returns D2D_TEST_SKIPPED in every mode (rc 77 checked), compare reads plain integers; oracle rerun on the rebased tree: claim 0 of 2,469 differ, parent 228, room control MB_SPOT=5 0, d2d claim output byte-identical to r1; body and trace fixed (three commits, mutants paragraph, four-class jump table); full ctest queued on the builder long lane (decomp44-monster-block-r2.md section 4); pets: no PR candidate (pet look also wrong in d2d, no oracle runs game.exe pet creation), D6a-d in decomp44-pet-block-trace.md; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Consumer: the coordinator's upstream decision on fid/monster-block-r2 (TASK-683, TASK-691).

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 r2 rebased on current upstream main; oracle, claim and parent compare rerun
- [x] #2 skip code fix and body fixes done
- [x] #3 pet block decomp written, candidate decision stated
<!-- AC:END -->
