---
id: TASK-813
title: 'necro4: Necromancer upkeep, curse, golem and corpse rules as classdata'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T19:08:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 372000
card: "necro4"
owns:
  - "internal/combat/classdata.go"
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/necro.go"
  - "internal/combat/necro_corpse.go"
  - "internal/combat/necro_test.go"
  - "internal/combat/necro_data_test.go"
  - "internal/combat/fighter.go"
exclusive: []
contracts: []
result: "Necro block in classdata profiles.json (golems, summons, curses with state and when, explosion); validated at load; necro.go reads it; build-specific data lives there; not verified live (offline tests, whole-repo green on the worker host)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: coordinator brief 2026-11-22; builds are data. Done when: no skill literal for summons, curses, golems or corpse explosion in necro.go.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Curse rules, curse states, golem ladder, summon list and corpse explosion name read from the profile necro block; no skill literal left in necro.go
- [x] #2 Unknown class or profile without the block is an error naming it; tests on Necromancer and Paladin
- [x] #3 Whole-repo vet and short tests green on the worker host
<!-- AC:END -->
