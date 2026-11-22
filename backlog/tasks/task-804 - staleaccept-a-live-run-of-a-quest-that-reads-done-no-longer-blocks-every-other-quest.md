---
id: TASK-804
title: 'staleaccept: a live run of a quest that reads done no longer blocks every other quest'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T15:37:00Z'
labels:
  - bot
  - bug
  - T0
dependencies: []
priority: high
ordinal: 369000
card: "staleaccept"
owns:
  - "internal/questboard/exec.go"
  - "internal/questboard/exec_sm.go"
  - "internal/questboard/exec_stale_test.go"
  - "tools/provingground/scenarios/board.json"
  - "docs/questboard.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User bug 2026-11-22: tower/countess stayed a live run after the quest word read done; 409 busy for every quest. Done when the load step closes it (auto-finish).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Unit test reproduces the stale accept and fails without the fix
- [ ] #2 Busy check table over quest done x chunk done x stopped on two classes
- [ ] #3 Scenario board.stale-accept-unblocks green on a lab slot
<!-- AC:END -->
