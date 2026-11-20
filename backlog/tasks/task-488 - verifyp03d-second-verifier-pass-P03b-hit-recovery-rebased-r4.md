---
id: TASK-488
title: 'verifyp03d: second verifier pass P03b hit recovery, rebased r4'
status: In Progress
assignee:
  - live-agent
created_date: '2026-11-15T09:51:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 271000
card: "verifyp03d"
owns:
  - "docs/upstream/verifier2-p03b-2026-11-15.md"
exclusive: []
contracts: []
result: "PASS WITH CHANGES (offline, not live): r4 a5bdd467 = r3 rebased on upstream fb6d799d, pushed to the private fork; builds, tests, compare 0 differ, mutant rc 134, ctest same as upstream, tidy 0, pr_guard 0/0. Required before posting: oracle needs EMU/GAMEEXE env (re-check fails), knock from a real get-hit still restarts the hold (FUN_00580a70 asks FUN_0057edd0)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: second independent check of upstream candidate P03b after #541 merged; the candidate must be rebased onto upstream main. Consumer: prfix3 (TASK-474), the posting step. Done when: the verdict doc names the head to post.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 verdict doc docs/upstream/verifier2-p03b-2026-11-15.md with rebased r4 head
- [x] #2 rebased branch built and tested on the NAS builder, pushed to the private fork only
<!-- AC:END -->
