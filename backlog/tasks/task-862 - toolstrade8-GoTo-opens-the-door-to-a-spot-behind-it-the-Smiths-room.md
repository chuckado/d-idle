---
id: TASK-862
title: 'toolstrade8: GoTo opens the door to a spot behind it (the Smith''s room)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T06:08:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 389000
card: "toolstrade8"
owns:
  - "internal/nav/nav.go"
  - "internal/nav/gatedoor.go"
  - "internal/nav/gatedoor_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "GoTo opens a door between the hero and an unreachable spot (up to 4 doors); unit-tested; lab run with the Smith pending."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. With quests.layoutRead 1 and a levelmonsters binary the Smith preset is found (754 at 23,57) but his room is behind closed doors: goto anchor reached only the closest point. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GoTo opens the door between the character and a spot a walk could not reach; must-fail test, two classes
<!-- AC:END -->
