---
id: TASK-453
title: 'uis4: Orders, read-only, as a slot panel'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T22:36:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "uis4"
owns:
  - "internal/watch/commands.go"
  - "internal/watch/commands_test.go"
  - "internal/watch/testdata/commands/"
  - "internal/watch/web/panels/orders/"
  - "internal/watch/testdata/orders_dom.js"
  - "internal/watch/testdata/orders_dom.txt"
  - "tools/uishots/fixtures_commands.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the Orders panel of the cockpit rail (J section 2, C-orders.md): one list of the player's commands in the bot's own order with one state vocabulary, instead of five queues in five places. Read-only: the verb row is drawn but inert (S5 adds the write side); the cancel cross posts to the route the row's own source already has. Done when: the criteria below are ticked and the result says what was verified. Based on agent/uis1b (S1b is not on main yet); one-line edits outside the owned paths are listed in the result.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Handler goldens testdata/commands/view-{empty,three-one-failed,blocked,paused,safety}-{sorceress,paladin}.json
- [ ] #2 A table test: every source state maps to exactly one row state
- [ ] #3 DOM goldens for the five fixtures; with eight rows the list scrolls inside and the fit check still passes at all 8 sizes
- [ ] #4 Cancel of an order, a queue entry, a repair and an item action each reaches its existing route (node test)
- [ ] #5 A race test over concurrent desk updates and reads passes under -race
<!-- AC:END -->
