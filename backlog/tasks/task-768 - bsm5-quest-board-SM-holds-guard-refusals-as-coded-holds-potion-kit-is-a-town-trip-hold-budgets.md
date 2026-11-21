---
id: TASK-768
title: 'bsm5: quest board SM holds (guard refusals as coded holds, potion kit is a town trip, hold budgets)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T05:06:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 344000
card: "bsm5"
owns:
  - "internal/task/gate.go"
  - "internal/task/kind.go"
  - "internal/task/executor.go"
  - "internal/task/world.go"
  - "internal/task/gate_hold_test.go"
  - "internal/bot/task.go"
exclusive: []
contracts: []
result: "With quests.boardSM=1: guards in task/gate.go return coded faults (needs_potions with a chores Insert, needs_gold_farm, too_weak, area_banned, route_closed); refusals are holds (Fault.Hold): the executor never counts or escapes them, the board shows the hold, the hero trains meanwhile (bot boardTrain) and the run parks with hold_timeout after the code budget. Knob 0 identical (faults keep no code). Unit tests on Paladin and Sorceress worlds with the knob-off contrast. Not verified live: scenarios board.hold.* stay pending (need a builder lane)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Card 5 of docs/design/fable-review-2026-11-21/quest-board.md section 5. Scope note: routeguard.go is not touched (route and tower reasons already arrive as World strings); the codes are attached in gate.go. U: area_banned/route_closed/too_weak are held and the hero trains meanwhile (boardTrain), needs a live scenario.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fault.Code/Hold and World.BoardSM: guards return coded holds only with quests.boardSM, knob 0 keeps the old faults
- [x] #2 potion kit refusal inserts the chores step (bug 4); a hold is never counted or escaped by the executor
- [x] #3 hold budgets park with hold_timeout; Paladin and Sorceress tests, contrast with knob off
<!-- AC:END -->
