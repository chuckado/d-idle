---
id: TASK-666
title: 'cainfix2: tree class from the preset, chunks stay done, hand-in waits for town'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-19T15:58:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 327000
card: "cainfix2"
owns:
  - "internal/d2dsrc/items.go"
  - "internal/d2dsrc/source_test.go"
  - "internal/bot/questboard.go"
  - "internal/bot/questboard_s7b_test.go"
  - "internal/bot/questboard_goal_test.go"
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
- [ ] #1 npcs-listed placed objects get their preset class (Tree of Inifuss class 30), must-fail test
- [ ] #2 item and auto chunks stay done once seen done, must-fail test
- [ ] #3 a hand-in tap waits in the queue until the hero is in town, must-fail test
<!-- AC:END -->
