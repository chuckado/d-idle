---
id: TASK-881
title: 'toolstrade15: GoTo plans through closed doors (live Barracks capture fixture)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T09:38:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 396000
card: "toolstrade15"
owns:
  - "internal/nav/gatedoor.go"
  - "internal/nav/gatedoor_test.go"
  - "internal/nav/testdata/barracks28.grid"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "doorOnPath: plan on a copy of the grid with every unused door open, open the first closed door on that path; unit-tested on the live Barracks grid; live effect not yet seen."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Live capture of the Barracks: the doors stand 22 subtiles from the dead end, outside the 14 subtile door reach. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GoTo to a spot behind a chain of closed doors plans through the doors, opens the first closed one on the path and walks on; must-fail test on the live Barracks grid, two classes
<!-- AC:END -->
