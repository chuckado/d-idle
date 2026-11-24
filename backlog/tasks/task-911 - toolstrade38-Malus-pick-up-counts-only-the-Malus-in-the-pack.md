---
id: TASK-911
title: 'toolstrade38: Malus pick-up counts only the Malus in the pack'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T00:59:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 416000
card: "toolstrade38"
owns:
  - "internal/bot/task.go"
  - "internal/bot/task_test.go"
  - "internal/bot/fakes_test.go"
  - "docs/probes/stepgrounds-stand.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 PickUp of a named quest item is success only when the item is in the pack; must-fail test
- [ ] #2 step.act1.smith.stand_click_pickup GREEN in the lab (or evidence why not)
<!-- AC:END -->
