---
id: TASK-819
title: 'necroscen: necro.* Proving Ground scenarios (upkeep, curse, revive, mana) with pre.spawn and minion/cast assertions'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T20:10:00Z'
labels:
  - bot
  - drill
dependencies: []
priority: medium
ordinal: 374000
card: "necroscen"
owns:
  - "tools/provingground/scenarios/necro.json"
  - "internal/scenario/necro_test.go"
  - "internal/scenario/setup.go"
  - "internal/scenario/state.go"
  - "internal/scenario/scenario.go"
  - "cmd/d2scenario/run.go"
exclusive: []
contracts: []
result: "Added pre.spawn (debug spawn), assertions minions_at_least (debug minions) and hero_casts_at_least (debug combatlog + skillstate names), runner needs read.* caps; scenarios necro.upkeep-pack, curse-unique, revive-after-death, mana-budget-10min and Paladin pair necro.paladin-pack in tools/provingground/scenarios/necro.json. Offline validated only; builder run pending."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 pre.spawn, minions_at_least and hero_casts_at_least exist with must-fail tests
- [x] #2 Four Necromancer scenarios and a Paladin pair validate
- [ ] #3 Run on the builder LONG lane with the lab d2d that has read.minions (devp20); RED/GREEN verdict recorded
<!-- AC:END -->
