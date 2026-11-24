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
  - "internal/nav/gatedoor.go"
  - "tools/probes/linkprobe.py"
  - "tools/probes/crossprobe.py"
  - "tools/probes/edgeprobe.py"
  - "tools/probes/crossmove.py"
  - "tools/probes/nearobj.py"
exclusive: []
contracts: []
result: "Cause (V lab devp27 objectlist and seeds 3, 7, 11): the Cathedral entrance of the Inner Cloister is a closed door in a notch of the north wall (class 25 Closed Door at tile (10.3, 0.1), rows 0 and 1 blocked above the one open notch row); d2d does not link the neighbour until it is open; the unproven-strip push never used it. The lab seed 11 crossed by luck (the Paladin Charge hit a monster behind the wall, 200 ms after the first push), seeds 3 and 7 never did (RED), which is what the live attempt 3 saw. Fix: openDoorAtGate opens the nearest closed door within 8 subtiles of an unproven border gate before pushing; the unproven strip is a frontier over 12 rows (segments by depth, the shallowest tried first, a notch is a gate of its own, the push comes from the neighbour edge), bad gates rotate, the first gate is lingered at for 75 s with the level read afresh, and the scan is logged (nav: crossing an unproven border strip ... scan=, nav: lingering at an unproven border ...). Lab nav.inner-cloister-cathedral GREEN for Paladin seeds 3, 7, 11 and Necromancer seed 3 (devp27, CHEATED setup, job 20261124-143010-9dfc); RED on seeds 3 and 7 before the door fix (job 20261124-142538-bbd9). Unit tests that fail without the code: TestExitToOpensTheDoorThatShutsAnUnprovenBorderGate (both classes), TestGatesFromAFrontierPushFromTheNeighboursEdge, TestUnprovenFrontierKeepsEveryDepth, TestUnprovenStripLooksInwardPastAWallEdge. The linger itself is lab-verified only (no discriminating unit test)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Unproven strip scanned to 12 rows over every column (frontier segments), pushed from the neighbour edge, bad gates rotated, the first gate lingered at, scan logged; must-fail unit tests on two layouts
- [x] #2 Lab nav.inner-cloister-cathedral on seeds 11, 3, 7
<!-- AC:END -->
