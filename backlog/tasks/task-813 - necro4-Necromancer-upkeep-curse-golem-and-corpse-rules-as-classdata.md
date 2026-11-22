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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Curse rules, curse states, golem ladder, summon list and corpse explosion name read from the profile necro block; no skill literal left in necro.go
- [ ] #2 Unknown class or profile without the block is an error naming it; tests on Necromancer and Paladin
- [ ] #3 Whole-repo vet and short tests green on the worker host
<!-- AC:END -->
