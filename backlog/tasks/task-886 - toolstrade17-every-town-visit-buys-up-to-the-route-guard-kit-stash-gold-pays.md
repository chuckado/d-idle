---
id: TASK-886
title: 'toolstrade17: every town visit buys up to the route guard kit, stash gold pays'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T10:39:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 398000
card: "toolstrade17"
owns:
  - "internal/bot/loop.go"
  - "internal/bot/beltfill_test.go"
  - "tools/provingground/scenarios/act1more.json"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "The loop town visits now ask the shop for the guard kit (belt plus pack), so shopCost includes it and the default withdraw-for-the-shop takes it from the stash; unit-tested, lab run queued, live effect not seen."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Live sandbox: the loop's own town visits bought portal scrolls and no healing; potion plan wantHP 0 with 2 healing carried against a floor of 4. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fullChores sets KitHealing and KitMana to the guard floor plus buffer (guard on, Normal, kit gate level); must-fail test, two classes
- [x] #2 Scenario act1.tools.malus.basic starts poor with a rich stash
<!-- AC:END -->
