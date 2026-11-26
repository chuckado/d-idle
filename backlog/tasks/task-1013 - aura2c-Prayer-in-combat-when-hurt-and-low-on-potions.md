---
id: TASK-1013
title: 'aura2c: Prayer in combat when hurt and low on potions'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T18:34:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 485000
card: "aura2c"
lane: next
owns:
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/classdata.go"
  - "internal/combat/aura.go"
  - "internal/combat/fighter.go"
  - "internal/combat/fightprayer_test.go"
exclusive: []
result: "Unit-tested only, not live. Prayer is wanted in a fight below 50 percent life with 2 or fewer healing potions in the belt (data row, both Paladin profiles), kept until 65 percent. Builds on aura2a."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: Prayer was resting-only; a Paladin out of potions and hurt in a fight kept Might. Done when: data row (fightLifePct 50, maxPotions 2) on Prayer in both Paladin profiles, with hysteresis, only on a known belt.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Both Paladin profiles carry a fightLifePct and maxPotions row on Prayer; wantAura honours it with hysteresis
- [x] #2 Test for both Paladin profiles
<!-- AC:END -->
