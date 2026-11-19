---
id: TASK-619
title: 'questrun: Start runs under a farm goal, queued reason, board log lines'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T14:59:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 310000
card: "questrun"
owns:
  - "internal/bot/questboard.go"
  - "internal/bot/questboard_goal_test.go"
  - "internal/bot/goal.go"
  - "internal/questboard/desk.go"
  - "internal/questboard/desk_test.go"
  - "internal/questboard/exec.go"
  - "cmd/d2bot/questwire.go"
  - "docs/questboard.md"
  - "internal/bot/routeguard.go"
  - "internal/bot/ladder_t0_test.go"
  - "internal/watch/web/panels/needs-you/panel.js"
  - "internal/watch/web/panels/needs-you/panel.css"
exclusive: []
contracts: []
result: "Cause: boardStep only reachable under a campaign goal. Fixed: boardPreempts, queued reason, quest board log lines; offline tests on two classes; Also: the tower ask (t0.area.toohard) is no longer raised when the hero is at or above the floor; Needs-you question text expands on tap. NOT verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: the board ran only under a campaign goal; a Start tap under farm_area was recorded and ignored (live 2026-11-17). Done when: the chunk preempts farm/level/boss goals at a safe moment and shows a queued reason.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Accepted chunk starts under a farm_area/reach_level goal at a safe moment
- [x] #2 Queued reason in strip/API and INFO log lines for board events
- [x] #3 Tests on two classes, knob 0 unchanged; docs/questboard.md updated
<!-- AC:END -->
