---
id: TASK-748
title: 'town2: supply.CountOf, the one counter, five old counters call it'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
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
  - "internal/bot/kitgate.go"
  - "internal/town/basicsfix.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 2 (B5). Knob-free; behaviour change only where the old counters disagreed, listed in the result.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 internal/supply Count types and CountOf, pure, belt plus pack for healing and mana, below_floor separate
- [ ] #2 beltAndInv, resInput belt loop, kitRead, healingCarried call it; fixture test lists where old and new differed
- [ ] #3 must-fail tests on Paladin and Sorceress; no class or skill literal in internal/supply (lint test)
<!-- AC:END -->
