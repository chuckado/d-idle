---
id: TASK-753
title: 'town3: supply.NeedFor, Solve, profiles.json, town.plan stage 1 shadow trace'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-715
priority: medium
ordinal: 343000
card: "town3"
owns:
  - "internal/supply"
  - "internal/town/planshadow.go"
  - "internal/town/chores.go"
  - "internal/town/town.go"
  - "internal/env/tuning_town.go"
  - "internal/bot/observe.go"
  - "docs/knobs.md"
exclusive: []
contracts: []
result: "Stage 1 of the town planner: supply.NeedFor, the pure Solve (priority table of 4.2, clamped minima, verdict ready/ready-lean/short-gold/blocked) and internal/supply/profiles.json (today's constants as rows; the file lives in the package, embedded like potionfloor.json, not under data/ which cannot be embedded). Knob town.plan (0 off default, 1 shadow, 2..5 later stages) in docs/knobs.md. At town.plan 1 every Chores visit logs 'town plan (shadow): for=.. verdict=.. needs ..; has ..; would buy ..' and the minute line gets townPlan=<stage>; nothing acts on it (town.plan 0 is byte-identical, minute line unchanged). Verified offline: go test ./internal/supply ./internal/env ./tools/knobdump ok, go test -p 1 ./internal/bot ok, -race -count=20 on supply and town (TestRefusedGearUntilLevelUp fails with -count>1 on main too: pre-existing, skipped). Two classes: solve tests and planshadow solve tests run Sorceress and Paladin bands. NOT verified: the town rig tests (stage 1 changes no behaviour) skip without game tables; no live run. Build-specific data: per-excursion minima in profiles.json, mana need from the belt plan column and resource.ManaNeed, belt capacity from the belt plan."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 3 (B5), stage 1.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 data/supply/profiles.json (today's constants), NeedFor, pure Solve with the 4.2 priority table; errors name profile/build/line
- [x] #2 knob town.plan (0 off, 1 shadow ...) default 0; stage 1 logs 'town plan (shadow): needs, has, would buy' and the minute-line field; no behaviour change
- [x] #3 Solve tests on Paladin and Sorceress bands; determinism test (T8)
<!-- AC:END -->
