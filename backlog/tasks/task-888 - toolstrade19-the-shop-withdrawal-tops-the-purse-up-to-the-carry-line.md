---
id: TASK-888
title: 'toolstrade19: the shop withdrawal tops the purse up to the carry line'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T11:07:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 400000
card: "toolstrade19"
owns:
  - "internal/town/basicsfix.go"
  - "internal/town/purserig_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Withdraw for the shop tops the purse up to the carry line; test passes on the builder (tables); the must-fail run did not finish, reasoning only; live effect not seen."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Lab and live: each visit withdrew only the shopping list and the purse stayed at 100 to 190. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 withdraw for the shop takes the pack up to the carry line (300 + 50 per level), default path; test on two classes
<!-- AC:END -->
