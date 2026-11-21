---
id: TASK-748
title: 'town2: supply.CountOf, the one counter, five old counters call it'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
updated_date: '2026-11-21T06:01:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-715
priority: medium
ordinal: 342000
card: "town2"
owns:
  - "internal/supply"
  - "internal/bot/goal.go"
  - "internal/bot/resource.go"
  - "internal/bot/supplycount.go"
  - "internal/town/basicsfix.go"
exclusive: []
contracts: []
result: "internal/supply.CountOf is the one potion counter (belt plus pack, rejuvenation counted as healing as before, below_floor separate, best belt tier); the bot's beltAndInv, resInput and (through resInput) kitRead and the town's healingCarried now call it with NO tier floor, so stage 0 numbers are the old ones. Verified offline: go test -p 1 ./internal/supply ./internal/town ./internal/bot ok, -race -count=20 on supply. Differences found by reading, not changed at stage 0: (1) kitRead and World.ManaPotions count mana in the belt only, the counter Mana() counts belt plus pack (used from town.plan 2, card town4); (2) the bot classifies by IsA, the town by exact type (belt.PotionKind), bot tier exists for healing only; (3) resInput never counted pack mana. The town rig tests skip on every host without game tables, so the town side is covered by the unchanged old tests only. Not done: kitRead keeps its own mana read (belt-only on purpose)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 2 (B5). Knob-free; behaviour change only where the old counters disagreed, listed in the result.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/supply Count types and CountOf, pure, belt plus pack for healing and mana, below_floor separate
- [x] #2 beltAndInv, resInput belt loop, kitRead, healingCarried call it; the differences are listed in the result (a fixture test needs game tables this host lacks)
- [x] #3 must-fail tests on Paladin and Sorceress; no class or skill literal in internal/supply (lint test)
<!-- AC:END -->
