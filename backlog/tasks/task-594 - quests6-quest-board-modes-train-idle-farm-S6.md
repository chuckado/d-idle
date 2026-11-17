---
id: TASK-594
title: 'quests6: quest board modes train idle farm (S6)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T06:28:00Z'
updated_date: '2026-11-17T07:25:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 304000
card: "quests6"
owns:
  - "internal/questboard"
  - "docs/design/quest-board-2026-11-16.md"
  - "internal/bot/boardmode.go"
  - "internal/bot/boardmode_test.go"
  - "internal/bot/questboard.go"
  - "docs/questboard.md"
  - "docs/farmloop.md"
  - "internal/env/tuning_quests.go"
exclusive: []
contracts: []
result: "S6 built offline behind quests.board: Train (pinned next quest, meter words), Idle (parked in town, watchdog/stall/drought quiet), Farm (farm-gold rung zone choice); safe-moment adoption, chunk precedence, two-class tests; NOT verified live (checklist in docs/questboard.md Modes)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: between quests the player picks what the hero does (design 3.4, S6). Done when: merged behind quests.board, live 30 min per mode run later.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Modes in the loop: Train pins the next quest, Idle parks in town with the watchdogs quiet, Farm uses the farm-gold zone choice
- [x] #2 Mode changes adopted at a safe moment; chunk precedence, orders override, knob 0 unchanged
- [x] #3 Readiness meter in words in GET /api/quests (meter)
- [x] #4 Tests on two classes; docs Modes section and live checklist
<!-- AC:END -->
