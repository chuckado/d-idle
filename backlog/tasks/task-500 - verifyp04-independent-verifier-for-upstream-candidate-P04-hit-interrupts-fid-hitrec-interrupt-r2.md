---
id: TASK-500
title: 'verifyp04: independent verifier for upstream candidate P04 (hit interrupts, fid/hitrec-interrupt-r2)'
status: In Progress
assignee:
  - verifyp04
created_date: '2026-11-15T14:57:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 276000
card: "verifyp04"
owns:
  - "docs/upstream/verifier-p04-2026-11-15.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P04 (a hit interrupts a swing or cast whose skill allows it) needs an independent verifier before the user's push go. Done when: the verdict doc names the exact head to post. Consumer: the upstream PR queue (docs/upstream/pr-queue.md, row P04).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 fid/hitrec-interrupt-r2 rebased on current upstream main and pushed to the private fork only
- [ ] #2 oracle and claim rebuilt and re-run on the NAS builder; claim test fails with the fix reverted
- [ ] #3 game side re-derived from game.exe (FUN_0057eec0, FUN_005809d0, FUN_00580a70)
- [ ] #4 hygiene checked: tidy, presubmit, pr_guard, trailers, identity, body
- [ ] #5 verdict doc docs/upstream/verifier-p04-2026-11-15.md with required changes and the head to post
<!-- AC:END -->
