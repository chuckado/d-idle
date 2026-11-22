---
id: TASK-819
title: >-
  necroscen: necro.* Proving Ground scenarios (upkeep, curse, revive, mana) with
  pre.spawn and minion/cast assertions
status: Done
assignee:
  - credit-runner
created_date: '2026-11-22T20:10:00Z'
updated_date: '2026-11-22T22:27:00Z'
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
result: "Added pre.spawn (debug spawn), assertions minions_at_least (debug minions) and hero_casts_at_least (debug combatlog + skillstate names), runner needs read.* caps; scenarios necro.upkeep-pack, curse-unique, revive-after-death, mana-budget-10min and Paladin pair necro.paladin-pack in tools/provingground/scenarios/necro.json. Builder LONG lane run 20261122-170600 (lab d2d devp20, combat.minionRead 1): GREEN 5, RED 0, BLOCKED 0 (the first run showed 2 RED on assertions that were wrong: no Teeth in the SME table, Paladin goes to town; and Revive needs clvl 36). All setups CHEATED (position and spawn): mechanics only. The knob is a rollback of the census source, so no RED with it off is expected."
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
- [x] #3 Run on the builder LONG lane with the lab d2d that has read.minions (devp20); RED/GREEN verdict recorded
<!-- AC:END -->
