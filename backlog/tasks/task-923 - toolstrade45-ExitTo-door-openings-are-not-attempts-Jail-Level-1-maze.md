---
id: TASK-923
title: 'toolstrade45: ExitTo door openings are not attempts (Jail Level 1 maze)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T16:57:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 425000
card: "toolstrade45"
owns:
  - "internal/nav/exit.go"
  - "internal/nav/gatedoor_test.go"
exclusive: []
contracts: []
result: "Cause (V live log 2026-11-24, 16:46 to 16:47Z): Jail Level 1 shuts the stairs to Jail Level 2 behind eight doors in a row ((210,205), (175,220), (145,210), (140,175) twice, (130,200), (105,190), (90,150)); ExitTo allows eight attempts and every door opened used one, so the ninth step ended walking out to level 30 did not change the level. The next call succeeded (Entered Jail Level 2 at 16:48). A door opened no longer counts as an attempt (bounded by 24 doors and the call deadline). Reproduced by TestExitToCrossesTenDoorsInARow (failed after 8 doors), not by a lab run. Not fixed: after the town trip the way back to level 30 timed out (town: no way back to level 30, route deadline), see the report."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ExitTo crosses more than eight doors in one call (Jail Level 1, ten doors in the test, two classes)
<!-- AC:END -->
