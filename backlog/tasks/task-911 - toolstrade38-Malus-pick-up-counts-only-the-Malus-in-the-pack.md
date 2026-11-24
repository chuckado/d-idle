---
id: TASK-911
title: 'toolstrade38: Malus pick-up counts only the Malus in the pack'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T00:59:00Z'
updated_date: '2026-11-24T01:29:00Z'
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
result: "Root cause: the Malus step trusted the loot pick-up count (three other items taken), so the stand was never clicked; PickUp of a quest item now needs the item in the pack. step.act1.smith.stand_click_pickup GREEN in the lab (V run 20261123-200329, Paladin 11, devp25); asserts the click because the chunk hands the Malus in within the budget."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 PickUp of a named quest item is success only when the item is in the pack; must-fail test
- [x] #2 step.act1.smith.stand_click_pickup GREEN in the lab (or evidence why not)
<!-- AC:END -->
