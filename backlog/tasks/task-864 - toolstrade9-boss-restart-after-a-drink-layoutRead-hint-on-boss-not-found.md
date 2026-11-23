---
id: TASK-864
title: 'toolstrade9: boss restart after a drink, layoutRead hint on boss not found'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T06:35:00Z'
updated_date: '2026-11-23T07:07:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 390000
card: "toolstrade9"
owns:
  - "internal/bot/task.go"
  - "internal/bot/task_test.go"
  - "internal/boss/act1.go"
  - "internal/env/tuning_quests.go"
  - "docs/process/layout-read.md"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "A drink no longer delays the boss restart by 30 s; the not-found error names quests.layoutRead for the Smith and Countess; unit-tested, lab run to green pending."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Lab: the Smith run was cancelled by drink-life again and again and each restart waited 30 s in the crowd; boss not found was a mystery without the knob. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A drink that cancels a boss run does not start the 30 s restart gap; must-fail test
- [x] #2 Boss not found with quests.layoutRead 0 says needs quests.layoutRead 1 for named-anchor bosses; must-fail test; documented as a prerequisite
<!-- AC:END -->
