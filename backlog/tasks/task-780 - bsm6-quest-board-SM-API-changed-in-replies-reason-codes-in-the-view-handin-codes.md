---
id: TASK-780
title: 'bsm6: quest board SM API (changed in replies, reason codes in the view, handin codes)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T17:21:00Z'
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
- [ ] #1 Reply.changed lists the runs a verb touched; the view carries reason codes per chunk row; handin_ready offer
- [ ] #2 scenarios stop-cascades, view.reason-always, handin.field-tap, handin.without-run green live (the second worker)
<!-- AC:END -->
