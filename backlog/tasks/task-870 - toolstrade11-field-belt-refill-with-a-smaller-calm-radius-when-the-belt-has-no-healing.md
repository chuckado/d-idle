---
id: TASK-870
title: 'toolstrade11: field belt refill with a smaller calm radius when the belt has no healing'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T07:45:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 392000
card: "toolstrade11"
owns:
  - "internal/bot/resource.go"
  - "internal/bot/beltfill_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "The field belt fill gate shrinks from 30 to 12 subtiles when the belt has no healing and the pack has; unit-tested; the in-field effect on the Smith run is not measured."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Lab and live: belt empty with potions in the pack, no calm moment within the 30 subtile sweep radius in the Barracks. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 beltFillStep fills from the pack with hostiles beyond 12 subtiles when the belt holds no healing potion and the pack does; must-fail test, two classes
<!-- AC:END -->
