---
id: TASK-1023
title: 'act2temple: Act 2 quest dungeon life floors and crowd-on-arrival rule'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T20:00:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 486000
card: "act2temple"
lane: next
owns:
  - "internal/plandata/admission.json"
  - "docs/plan/data/admission.json"
  - "internal/admit/"
  - "internal/combat/arrival.go"
  - "internal/combat/arrival_test.go"
  - "internal/combat/safety.go"
  - "internal/combat/fighter.go"
  - "internal/combat/crowd.go"
exclusive: []
contracts: []
result: "see report"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Claw Viper Temple stand-in dies 1 s after arriving among 22 to 39 hostiles (matrix note 80). Done when the arrival row sends the hero back at once and the drill has a named outcome.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Road rows of the Act 2 quest dungeons carry level and life floors
- [x] #2 Arrival crowd rule as data, unit tested
- [ ] #3 drill.act2.amulet repeat 3 on the worker host
<!-- AC:END -->
