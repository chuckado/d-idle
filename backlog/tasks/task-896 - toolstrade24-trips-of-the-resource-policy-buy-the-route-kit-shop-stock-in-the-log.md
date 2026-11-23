---
id: TASK-896
title: >-
  toolstrade24: trips of the resource policy buy the route kit; shop stock in
  the log
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T12:45:00Z'
updated_date: '2026-11-23T13:05:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 404000
card: "toolstrade24"
owns:
  - "internal/bot/goal.go"
  - "internal/bot/loop.go"
  - "internal/bot/beltfill_test.go"
  - "internal/town/shop.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Trips of the resource policy now carry the route kit like fullChores and the task chores; the vendor stock is logged; unit-tested, live effect not seen; the Smith fight is not touched."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Live 12:41Z: gold 1160, potion plan wantHP 0, five portal scrolls bought, the guard blocked again. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A potion trip built by the resource policy gets the route guard kit (withRouteKit); must-fail test, two classes
- [x] #2 The shop logs the healing and mana potions the vendor holds
<!-- AC:END -->
