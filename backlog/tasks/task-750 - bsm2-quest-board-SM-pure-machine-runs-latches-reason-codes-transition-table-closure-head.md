---
id: TASK-750
title: 'bsm2: quest board SM pure machine (runs, latches, reason codes, transition table, closure, head)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 341000
card: "bsm2"
owns:
  - "internal/questboard/sm.go"
  - "internal/questboard/sm_reason.go"
  - "internal/questboard/sm_test.go"
exclusive: []
contracts: []
result: "Pure machine in internal/questboard/sm.go + sm_reason.go (not wired): run statuses, total Step table, closed reason table (B.1), Start/Stop/Retry/Promote/NoteDone/Regress/Check/Repair. Unit tests over every multi-part quest of the shipped board (stand-in for two builds: the model has no class), race count=20 green. Not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Card 2 of docs/design/fable-review-2026-11-21/quest-board.md section 5. Pure, not wired; Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Step is total over status x event, closed reason table with remedy and budget
- [x] #2 Start/Stop/Retry/Regress/Repair/Check unit tests (I1-I5, I7-I9, D1-D5, R1-R4) over every multi-part quest
- [x] #3 race count=20 on the questboard model tests
<!-- AC:END -->
