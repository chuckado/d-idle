---
id: TASK-852
title: 'toolstrade6: the next road step is ready for a hero on the road with a thinned kit'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T04:27:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 386000
card: "toolstrade6"
owns:
  - "internal/task/gate.go"
  - "internal/task/route_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "A hero already in a guarded area is ready for the next road step with 3 of 4 healing potions (halved floors, as guardedExit since toolstrade5); not verified in a lab run to green."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Live sandbox 2026-11-23: at the Outer Cloister the route guard (3 healing potions < 4) inserted a chores trip, the hero walked home through the crowd and the chunk restarted in a loop. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fieldReady lets a hero standing in a guarded area go on with a kit that thinned (RouteLeaveWhy); from town the gate is unchanged; must-fail test, two classes
<!-- AC:END -->
