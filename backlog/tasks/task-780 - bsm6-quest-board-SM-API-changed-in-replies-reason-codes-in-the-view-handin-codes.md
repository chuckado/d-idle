---
id: TASK-780
title: >-
  bsm6: quest board SM API (changed in replies, reason codes in the view, handin
  codes)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T17:21:00Z'
updated_date: '2026-11-21T18:40:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 360000
card: "bsm6"
owns:
  - "internal/questboard/desk.go"
  - "internal/questboard/desk_reason.go"
  - "internal/questboard/desk_reason_test.go"
  - "internal/questboard/exec_sm.go"
  - "internal/questboard/exec.go"
  - "internal/watch/questsapi.go"
  - "docs/questboard.md"
  - "docs/probes/boardsm-progress-2026-11-21.md"
  - "tools/provingground/scenarios/board.json"
  - "internal/scenario/board.go"
  - "internal/scenario/board_test.go"
  - "internal/scenario/state.go"
  - "internal/watch/testdata/quests/"
exclusive: []
contracts: []
result: "API: replies of start, cancel and retry carry changed rows (quest, chunk, state, reason); chunk rows and run views carry reason codes with sentences (waits_for, hold codes, parks, handin_wanted, confirming); handin_ready row reason, Start on it creates the awaiting run; assert board_reasons. Unit tested on Paladin and Sorceress, race x20. Live (second worker, CHEATED setups, mechanics only): stop-cascades GREEN Paladin and Sorceress, reason-always GREEN Sorceress (Paladin see the progress file); handin.field-tap and handin.without-run stay pending: the lab pending bit does not read as awaiting (U)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Reply.changed lists the runs a verb touched; the view carries reason codes per chunk row; handin_ready offer
- [x] #2 scenarios stop-cascades and view.reason-always green live (the second worker); handin.* stay pending (the setup does not read as awaiting, U)
<!-- AC:END -->
