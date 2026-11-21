---
id: TASK-755
title: 'town4: stage 2 guards, chunk preconditions and shop read the supply counter'
status: In Progress
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
  - "cmd/d2bot"
  - "docs/knobs.md"
exclusive: []
contracts: []
result: "Stage 2 of the town planner behind knob town.plan >= 2 (default 0: old behaviour, the minute line and the guards are unchanged). Guards: World.KitPlan (the supply plan's clamped minima for profile monastery_road and the line it cannot close) is read by kitWhyFor, guardFault and StepChores.Need; mana is counted belt plus pack; a kit-only refusal the gold cannot pay is a danger fault that names the line ('short of gold: needs N more for healing') instead of a chores trip that ends 'ran but is still needed'; a fixable one inserts the restock chores (chores2 mechanism, no failure counted). Shop: the task chores buy KitHealing and the new KitMana to the plan minima into the pack (belt plus pack counted); at stage 2 a belt-fill step runs before the sell chore. Verified offline: new must-fail tests on a Paladin and a Sorceress (internal/task kitplan_test.go, internal/bot kitplan_test.go), go test -p 1 ./internal/bot ./internal/task ./internal/town ./internal/supply ok, -race -count=20 on supply task and town (TestRefusedGearUntilLevelUp fails with -count>1 on main: pre-existing), cmd/d2bot wiring test ok (the merc town wrapper forwards SupplyPlan; TestBestiaryWiredWithTableAndRestart fails on main too). NOT verified: no Proving Ground run of guard-blocked-restock or mana-kit-satisfiable (offline twins only), the town rig tests (belt-first order, KitMana buy) skip without game tables, no live run. Not covered: the boss gates (BossGate kit) still use their own constants and the Andariel antidote line. Build data: minima from profiles.json, mana line only when the belt plan has a mana column."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 4 (B5), stage 2. Coordinates with the bsm cards (board Q9).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 guards (RouteKitWhy, TowerKitWhy, BossGate, StepChores.Need) read the verdict at town.plan>=2; block only on a line the plan cannot close and name it
- [x] #2 shop buys to the plan; belt fill before sell; a guard refusal fixable in town triggers an immediate restock trip, no failure counted
- [x] #3 offline twins of guard-blocked-restock and mana-kit-satisfiable pass (scenarios not run on the builder)
- [x] #4 -race -count=20 on touched packages; default stays 0 until green
<!-- AC:END -->
