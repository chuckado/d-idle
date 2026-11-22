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
result: "Fixed: Exec.smAutoFinish closes a live run whose chunk view reads done/skipped (quest word done), logged auto-finish; unit tests + table on 2 classes, -race x20 green; scenario board.stale-accept-unblocks GREEN on a lab slot (Paladin seed 11, CHEATED setup, mechanics only)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User bug 2026-11-22: forgotten_tower/countess stayed a live run after the quest word read done; 409 busy for every quest. Done when the load step closes it (auto-finish).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Unit test reproduces the stale accept and fails without the fix
- [x] #2 Busy check table over quest done x chunk done x stopped on two classes
- [x] #3 Scenario board.stale-accept-unblocks green on a lab slot
<!-- AC:END -->
