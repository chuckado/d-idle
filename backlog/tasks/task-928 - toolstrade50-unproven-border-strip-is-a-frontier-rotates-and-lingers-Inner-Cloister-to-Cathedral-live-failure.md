---
id: TASK-928
title: 'toolstrade50: unproven border strip is a frontier, rotates and lingers (Inner Cloister to Cathedral live failure)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T19:21:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 430000
card: "toolstrade50"
owns:
  - "internal/nav/exit.go"
  - "internal/nav/gatedoor_test.go"
  - "internal/d2dworld/border.go"
  - "internal/d2dworld/world.go"
  - "internal/d2dworld/world_test.go"
  - "internal/d2map/tracker.go"
  - "tools/provingground/scenarios/navexit.json"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Unproven strip scanned to 12 rows over every column (frontier segments), pushed from the neighbour edge, bad gates rotated, the first gate lingered at, scan logged; must-fail unit tests on two layouts
- [ ] #2 Lab nav.inner-cloister-cathedral on seeds 11, 3, 7
<!-- AC:END -->
