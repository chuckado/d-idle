---
id: TASK-885
title: 'toolstrade16: no home loop for a thinned kit anywhere out of town'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T10:25:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 397000
card: "toolstrade16"
owns:
  - "internal/task/gate.go"
  - "internal/task/route_test.go"
  - "tools/provingground/scenarios/act1more.json"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "The mid-route kit tolerance applies in every area outside town (it was only in guarded areas); unit-tested; live effect not seen; the TP scroll spend on a thin purse is an open finding."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Live deploy 46: the hero walked town, Cold Plains, Stony Field, Underground Passage, Dark Wood, Black Marsh and went home again at 3 healing potions of 4 before ever reaching the Barracks. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fieldReady lets a hero that is out of town go on with 3 of 4 healing potions in any area, not only in guarded ones; must-fail test, two classes
- [x] #2 Scenario act1.tools.malus.basic (no belt, 4 boxes) exists
<!-- AC:END -->
