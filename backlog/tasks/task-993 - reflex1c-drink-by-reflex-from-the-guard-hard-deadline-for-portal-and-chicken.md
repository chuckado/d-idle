---
id: TASK-993
title: 'reflex1c: drink by reflex from the guard, hard deadline for portal and chicken'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T04:07:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 457000
card: "reflex1c"
owns:
  - "internal/bot/safetylat.go"
  - "internal/bot/reflexdrink_test.go"
  - "internal/bot/loop.go"
  - "internal/combat/reflex.go"
  - "internal/combat/reflex_test.go"
  - "internal/combat/fighter.go"
  - "internal/combat/hard.go"
  - "internal/combat/packdrink.go"
  - "internal/combat/potioneffect.go"
  - "internal/env/tuning_bot.go"
  - "docs/bot.md"
  - "docs/safety-latency.md"
exclusive: []
result: "bot.reflexDrink (default 1): the guard sends a belt life or rejuvenation drink itself through the fighter (ReflexDrink, drink state behind a mutex) without cancelling the action; bot.reflexDeadlineMs (default 1000): a portal, chicken or crowd cancel the action ignores is carried out by the guard, once. Two classes in every test; mutation-checked. Offline only, not verified live; measure with safetyLatMaxMs and reflexDrinks on the minute line."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Fable review P3 (c); owner decision 2026-11-25.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 a belt drink is sent by the guard without cancelling the action (knob bot.reflexDrink, default 1); the pack path and portal, chicken, crowd still cancel
- [x] #2 after bot.reflexDeadlineMs the guard carries a hard cancel out itself, once
- [x] #3 docs/bot.md invariant amended
<!-- AC:END -->
