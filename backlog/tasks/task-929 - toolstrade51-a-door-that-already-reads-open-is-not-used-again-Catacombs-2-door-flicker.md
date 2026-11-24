---
id: TASK-929
title: 'toolstrade51: a door that already reads open is not used again (Catacombs 2 door flicker)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T19:59:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 431000
card: "toolstrade51"
owns:
  - "internal/nav/gatedoor.go"
  - "internal/nav/gatedoor_test.go"
exclusive: []
contracts: []
result: "Cause (V live capture catacombs2-1456 and nav lines 19:54Z, reproduced on the real grid in the sim): the way from the stairs of Catacombs Level 2 to the gate at (149,278) is two doors in a row, (200,210) in the corridor wall and then (190,220) west of the room below. The nav tried (190,220) first, twice, on walks that could not reach it, which used both of its maxDoorTries; after the door (200,210) opened it was never a candidate again, so the nav used the open door (200,210) a second time (cellFlags 0 then 1, it shut again), then tried (160,210) which is in another compartment, stepped to the nearest connected cell and gave up (three steps, parked gave_up). Fix: (1) the tries other doors spent on walks that could not reach them are forgiven once a door has opened (nav: door used with the door cell open), (2) a door whose own cell already reads walkable is no candidate (it would be shut again). Tests: TestExitToCrossesTheLiveCatacombsTwo on the live grid (testdata/catacombs35.grid, a crop of debug grid 5 with the doors shut as 3 by 3 blocks, both classes; fails without (1)), TestExitToDoesNotUseADoorThatIsAlreadyOpen (two layouts, fails without (2)). Whether the live way has more doors after (190,220) is U until attempt 5."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 An open door is not a candidate: unit test on two layouts fails without the check
<!-- AC:END -->
