---
id: TASK-755
title: 'town4: stage 2 guards, chunk preconditions and shop read the supply counter'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:51:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-715
priority: medium
ordinal: 344000
card: "town4"
owns:
  - "internal/task"
  - "internal/town"
  - "internal/bot"
  - "internal/supply"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 4 (B5), stage 2. Coordinates with the bsm cards (board Q9).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 guards (RouteKitWhy, TowerKitWhy, BossGate, StepChores.Need) read the verdict at town.plan>=2; block only on a line the plan cannot close and name it
- [ ] #2 shop buys to the plan; belt fill before sell; a guard refusal fixable in town triggers an immediate restock trip, no failure counted
- [ ] #3 scenarios guard-blocked-restock and mana-kit-satisfiable go GREEN (or offline twins if the builder is busy)
- [ ] #4 -race -count=20 on touched packages; default stays 0 until green
<!-- AC:END -->
