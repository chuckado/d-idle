---
id: TASK-897
title: 'toolstrade25: chickens count against their own limit, not the three of a death'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T13:05:00Z'
updated_date: '2026-11-26T01:45:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 405000
card: "toolstrade25"
owns:
  - "internal/task/kind.go"
  - "internal/task/executor.go"
  - "internal/task/executor_unpark_test.go"
  - "internal/bot/task.go"
  - "internal/bot/task_test.go"
  - "internal/bot/zz_cover_task_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Chickens park a quest after 8 in 90 minutes instead of 3; deaths stay at 3; unit-tested, live effect not seen; the Smith fight itself is unchanged."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Lab and live: three chickens at the Smith in a minute parked the chunk as too weak. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A chicken exit of a task step is KindChicken and counts against Policy.MaxChickens (8) apart from deaths (3); must-fail tests
<!-- AC:END -->
