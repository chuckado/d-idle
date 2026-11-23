---
id: TASK-904
title: 'toolstrade30: the Barracks life floor is waived four levels above the level floor'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T22:37:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 410000
card: "toolstrade30"
owns:
  - "internal/env/tuning_route.go"
  - "internal/task/route.go"
  - "internal/task/world.go"
  - "internal/task/route_test.go"
  - "internal/bot/routeguard.go"
  - "docs/knobs.md"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Life floor waived 4 levels above the level floor, logged once; hold text life X/Y; unit-tested; not live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Gate calibration risk from the Necromancer lab and the live Paladin. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The life floor of the Barracks stops applying at the level floor plus route.barracksLifeWaiver (4), is logged, and shows life X/Y in the hold text; must-fail test, Paladin and Necromancer numbers
<!-- AC:END -->
