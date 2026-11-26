---
id: TASK-902
title: >-
  toolstrade29: the farm walk keeps out of the Barracks below its floor; plan
  rows for clvl 15 to 24
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T22:10:00Z'
updated_date: '2026-11-26T01:45:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 409000
card: "toolstrade29"
owns:
  - "internal/bot/routeguard.go"
  - "internal/bot/plan_test.go"
  - "internal/bot/beltfill_test.go"
  - "internal/task/route.go"
  - "internal/task/gate.go"
  - "docs/plan/data/levelplan-normal.json"
  - "internal/plandata/levelplan-normal.json"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "The farm picker and the walk guard use the Barracks floor; the Paladin plan gets Tower Cellars rows for 15 to 24; unit-tested; xp effect not measured live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Live 2026-11-23: the farm walk to the Catacombs crossed the Barracks at clvl 20 (7 town portals and 6 chickens in area 28) and the plan's windows ended at 18 (28 XP/min). Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Areas 28 to 37 are excluded from the picker and refused to a walk below the class floor of the Barracks; must-fail test, two classes
- [x] #2 The Paladin level plan has a Tower Cellars slice for clvl 15 to 24; must-fail pick test
<!-- AC:END -->
