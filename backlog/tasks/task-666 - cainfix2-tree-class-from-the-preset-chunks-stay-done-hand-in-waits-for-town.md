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
  - "internal/task/step.go"
  - "internal/task/act1.go"
exclusive: []
contracts: []
result: "Verified live on a d2d lab slot (not Game.exe) by drill cainl3 (docs/probes/search-for-cain-l3-2026-11-19.md): the real bot chain finished the quest unattended with these fixes. Four fixes, each with a must-fail test: npcs-listed placed objects get their preset class (tree class 30, internal/d2dsrc); item and auto chunks stay done once seen done and count for later chunks; a board hand-in tap waits in the queue until the hero is in town; the free-Cain chunk is done on the quest word, not on standing outside Tristram (Step.Evidence). Class-free (no skill or build data). Not fixed: the default crowd rule chickens before the tree on the lab seed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 npcs-listed placed objects get their preset class (Tree of Inifuss class 30), must-fail test
- [x] #2 item and auto chunks stay done once seen done, must-fail test
- [x] #3 a hand-in tap waits in the queue until the hero is in town, must-fail test
- [x] #4 the free-Cain chunk is done on the quest word, not outside Tristram, must-fail test
<!-- AC:END -->
