---
id: TASK-541
title: 'idtome: Tome of Identify buy chore (off by default), docs'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T04:00:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 286000
card: "idtome"
result: "Optional idtome chore behind town.identifyTome (default 0, bot never auto-buys); the manual Buy scrolls errand already buys the tome and fills it; pickup stops at a full tome (existing). Tests skip without extracted tables; not verified live."
owns:
  - "internal/town/idtome_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 chore behind town.identifyTome default 0, manual buy unaffected
- [x] #2 tests on Sorceress and Paladin
<!-- AC:END -->
