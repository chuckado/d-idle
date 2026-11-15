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
result: "PASS (not verified live; offline: emulator on the worker host, NAS builder). Post fid/hitrec-interrupt-r3 5fa67510: all five required changes in; parent 2,811 differ, claim 0; three mutants abort; ctest identical to upstream; tidy clean in PR files; clean on upstream 1023c4ac. docs/upstream/verifier2-p04-2026-11-15.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: second independent pass on P04 after the fix agent's r3 (prfix7). Done when: the verdict names the head to post. Consumer: the upstream PR queue (docs/upstream/pr-queue.md, row P04).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 all five required changes of verifier-p04 checked on r3
- [x] #2 r3 rebuilt and re-run on the NAS builder with and without each fix; compare against a fresh game.exe grid
- [x] #3 state leaks checked: which unit states the gate reads and which d2d models
- [x] #4 rebase on current upstream main, ctest, clang-tidy, presubmit, pr_guard, trailers
- [x] #5 verdict doc docs/upstream/verifier2-p04-2026-11-15.md names the head to post
<!-- AC:END -->
