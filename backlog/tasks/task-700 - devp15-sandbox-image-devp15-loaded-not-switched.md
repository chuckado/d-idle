---
id: TASK-700
title: 'devp15: sandbox image devp15 loaded (not switched)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T19:57:00Z'
labels:
  - docs
  - chore
dependencies: []
priority: medium
ordinal: 246000
card: "devp15"
owns:
  - "docs/process/sandbox-devp15.md"
exclusive: []
contracts: []
result: "Image d2dbot:devp15 (code tip 5dacc604, test-fix tip 373b93e83) built and loaded on the worker host via the fork-rebase routine; smoke_d2d still fails at line 512 on devp13, devp14 and devp15 alike (pre-existing). Sandbox container untouched (V: docker ps after load). Image smoke by --version, no game run; not verified live in a game."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 image d2dbot:devp15 built from the combined fork branch, tests rc 0
- [x] #2 image loaded on the worker host, sandbox container unchanged
- [x] #3 docs/process/sandbox-devp15.md written (recreate and rollback notes)
<!-- AC:END -->
