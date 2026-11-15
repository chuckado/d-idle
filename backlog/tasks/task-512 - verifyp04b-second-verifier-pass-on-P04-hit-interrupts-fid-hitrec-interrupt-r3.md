---
id: TASK-512
title: 'verifyp04b: second verifier pass on P04 hit interrupts (fid/hitrec-interrupt-r3)'
status: In Progress
assignee:
  - live-agent
created_date: '2026-11-15T17:18:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 278000
card: "verifyp04b"
owns:
  - "docs/upstream/verifier2-p04-2026-11-15.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: second independent pass on P04 after the fix agent's r3 (prfix7). Done when: the verdict names the head to post. Consumer: the upstream PR queue (docs/upstream/pr-queue.md, row P04).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 all five required changes of verifier-p04 checked on r3
- [ ] #2 r3 rebuilt and re-run on the NAS builder with and without each fix; compare against a fresh game.exe grid
- [ ] #3 state leaks checked: which unit states the gate reads and which d2d models
- [ ] #4 rebase on current upstream main, ctest, clang-tidy, presubmit, pr_guard, trailers
- [ ] #5 verdict doc docs/upstream/verifier2-p04-2026-11-15.md names the head to post
<!-- AC:END -->
