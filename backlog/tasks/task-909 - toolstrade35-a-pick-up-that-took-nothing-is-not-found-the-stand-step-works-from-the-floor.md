---
id: TASK-909
title: >-
  toolstrade35: a pick-up that took nothing is not found; the stand step works
  from the floor
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T23:50:00Z'
updated_date: '2026-11-24T00:12:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 414000
card: "toolstrade35"
owns:
  - "internal/bot/task.go"
  - "internal/bot/task_test.go"
  - "internal/task/act1.go"
  - "internal/task/steps.go"
  - "internal/task/gear_gate_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "PickUp reports nothing taken as not found; the Malus step picks up first, clicks the stand when nothing lies there, and is ready in the Barracks; unit-tested; the stepgrounds scenarios are not run from here."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Stepgrounds runs 175729 and 175828: a Malus on the Barracks floor was not picked up in the chunk path. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 taskActs.PickUp returns KindNotFound when nothing was picked; the Malus step takes a Malus lying on the floor first, clicks the stand only when none lies there, and is ready in the Barracks without a kill in this game; must-fail tests
<!-- AC:END -->
