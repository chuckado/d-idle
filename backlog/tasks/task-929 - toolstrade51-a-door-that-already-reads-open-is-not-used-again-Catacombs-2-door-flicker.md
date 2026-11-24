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
result: "Cause (V live 19:54Z to 19:56Z, deploy 64, Catacombs Level 2): the corridor doors at (200,210), (160,210), (140,210) (class 16) and (190,220), (130,220) (class 64) are listed; the nav used the door at (200,210) (cellFlags 0 = open), still found the gate cut off by a barrier the map does not list as a door, picked the same door again (maxDoorTries 2) and shut it (cellFlags 1), then tried the door at (160,210) it cannot reach, stepped to the nearest connected cell and gave up after three steps. A door whose own cell already reads walkable is no candidate now (doorAlreadyOpen, in openDoorToward and openDoorAtGate), so the next door on the way (190,220) is tried instead. Unit test TestExitToDoesNotUseADoorThatIsAlreadyOpen, two layouts, used the open door twice before the change. Not reproduced in the lab (the live layout differs from seeds 3, 7 and 11) and the live capture of Catacombs 2 came back empty, so whether the door at (190,220) opens the way is U until attempt 5; the log lines to read are nav: door used (door, cellFlags) and nav: the gate is cut off by a door."
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
