---
id: TASK-836
title: 'toolstrade2: walks keep off border strips of unwanted levels; tools scenario hero fits the level'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T02:23:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 381000
card: "toolstrade2"
owns:
  - "internal/nav/nav.go"
  - "internal/nav/gatedoor_test.go"
  - "tools/provingground/scenarios/act1more.json"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Planner penalises border strips toward other levels (the Monastery Gate walk to a door ended in the Tamoe Highland); unit-tested only so far, scenario run pending."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Follow-up of toolstrade: re-entry from town ended in Tamoe Highland. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Planner puts a soft cost on border strips toward levels the walk does not want; must-fail test, two classes
- [x] #2 act1.tools.malus hero fits the level (clvl 19, merc, worn kit), CHEATED mechanics only
<!-- AC:END -->
