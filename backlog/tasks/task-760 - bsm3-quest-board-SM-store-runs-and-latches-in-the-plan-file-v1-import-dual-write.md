---
id: TASK-760
title: 'bsm3: quest board SM store (runs and latches in the plan file, v1 import, dual write)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:52:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 342000
card: "bsm3"
owns:
  - "internal/questboard/store_sm.go"
  - "internal/questboard/store_sm_test.go"
  - "internal/questboard/plan.go"
exclusive: []
contracts: []
result: "store_sm.go: LoadSM/UpdateSM reconcile the runs with Accepted (v1 import, withdrawn records stopped, D4 repair), running/held not stored, Accepted dual-written, stopped runs bounded. File version stays 1 (additive fields; an older build ignores them). Not wired (card 4). Not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Card 3 of docs/design/fable-review-2026-11-21/quest-board.md section 5. Decision to note: the file version stays 1 (additive fields) because a v2 file would be refused by an older build. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Plan holds runs and latches additively in file version 1; Accepted dual-written from live and finished runs
- [x] #2 v1 import of five shapes, D4 repair at load, withdrawn-record reconcile, round trip
- [x] #3 an older build still reads the file; race count=20 on the store tests
<!-- AC:END -->
