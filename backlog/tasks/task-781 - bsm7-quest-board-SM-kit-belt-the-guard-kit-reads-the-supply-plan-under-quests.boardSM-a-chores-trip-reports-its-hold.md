---
id: TASK-781
title: 'bsm7: quest board SM kit-belt (the guard kit reads the supply plan under quests.boardSM, a chores trip reports its hold)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T17:46:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 360000
card: "bsm7"
owns:
  - "internal/task/executor.go"
  - "internal/task/gate_hold_test.go"
  - "internal/bot/routeguard.go"
  - "internal/bot/questboard.go"
  - "internal/bot/kitplan_test.go"
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
- [ ] #1 with quests.boardSM the guards and the shop read the supply plan (belt plus pack, minima clamped to the belt) so a chores trip can close the kit
- [ ] #2 a coded fault that inserts the chores trip reports its hold (Result.Via) and the board logs 'quest board: hold needs_potions' with nothing counted
- [ ] #3 board.hold.* scenarios green live (Orion)
<!-- AC:END -->
