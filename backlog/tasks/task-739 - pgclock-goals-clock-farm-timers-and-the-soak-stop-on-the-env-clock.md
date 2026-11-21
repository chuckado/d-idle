---
id: TASK-739
title: 'pgclock: goals clock, farm timers and the soak stop on the env clock'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:47:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 345000
card: "pgclock"
owns:
  - "cmd/d2bot/d2dbackend.go"
  - "internal/bot/farm.go"
  - "internal/goals/clock_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design 2.3 B1 to B3, 7 card 5. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 goals.Manager gets the env clock on the d2d path
- [ ] #2 farm budget uses the env clock
- [ ] #3 soak stop is tick based
- [ ] #4 tests for two builds
<!-- AC:END -->
