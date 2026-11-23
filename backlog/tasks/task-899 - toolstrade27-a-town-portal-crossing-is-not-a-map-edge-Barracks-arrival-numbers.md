---
id: TASK-899
title: 'toolstrade27: a town portal crossing is not a map edge; Barracks arrival numbers'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T20:46:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 407000
card: "toolstrade27"
owns:
  - "internal/nav/route.go"
  - "internal/nav/gatedoor_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "learn() no longer records a town portal crossing as an edge; unit-tested; live effect not seen; the Barracks survival analysis is in the probe doc."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Live 20:13Z: the chunk parked on no exit to level 27 from level 1 after a chicken. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 learn ignores a crossing that involves a town unless the pair is a hint or levels.txt link; must-fail test, two classes
<!-- AC:END -->
