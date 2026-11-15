---
id: TASK-500
title: 'verifyp04: independent verifier for upstream candidate P04 (hit interrupts, fid/hitrec-interrupt-r2)'
status: In Progress
assignee:
  - live-agent
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
result: "PASS WITH CHANGES (not verified live; offline: emulator on the worker host, NAS builder). Reproduced 1,741/2,856 before, 0 after on the rebase fid/hitrec-interrupt-r2 98d4a942 (pushed to the private fork); mutant fails. Required for -r3: refuse in state 0x36 (Leap Attack regression), barrel kick as Kick, sequences.hpp include, state-branch and scroll wording, Checklist. docs/upstream/verifier-p04-2026-11-15.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P04 (a hit interrupts a swing or cast whose skill allows it) needs an independent verifier before the user's push go. Done when: the verdict doc names the exact head to post. Consumer: the upstream PR queue (docs/upstream/pr-queue.md, row P04).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fid/hitrec-interrupt-r2 rebased on current upstream main and pushed to the private fork only
- [x] #2 oracle and claim rebuilt and re-run on the NAS builder; claim test fails with the fix reverted
- [x] #3 game side re-derived from game.exe (FUN_0057eec0, FUN_005809d0, FUN_00580a70)
- [x] #4 hygiene checked: tidy, presubmit, pr_guard, trailers, identity, body
- [x] #5 verdict doc docs/upstream/verifier-p04-2026-11-15.md with required changes and the head to post
<!-- AC:END -->
