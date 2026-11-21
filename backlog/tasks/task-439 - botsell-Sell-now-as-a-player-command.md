---
id: TASK-439
title: 'botsell: Sell now as a player command'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T20:54:00Z'
updated_date: '2026-11-14T20:55:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 258000
card: "botsell"
owns:
  - "internal/invdesk/sellnow.go"
  - "internal/invdesk/sellnow_test.go"
  - "internal/watch/invapi_sellnow.go"
  - "internal/watch/invapi_sellnow_test.go"
  - "cmd/d2bot/invwire.go (your lines only)"
  - "docs/inventory-actions.md (a new section)"
exclusive: []
contracts: []
result: "built, NOT verified live (2026-11-14): invdesk kind sell (errand like Cain, only player-marked items, refuses with 'nothing is marked for sale'), POST /api/inventory/sellnow, view schema 10 (additive), town sale via Town.Chores with an empty plan. Fake tests for Paladin and Sorceress, route and view goldens, live checklist 21 to 27 in docs/inventory-actions.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: marked items sell only on the next town visit; the player wants a command that forces it (C-orders sell_now). Done when: the four boxes hold.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fake tests for both classes: in the field it goes home at a safe moment and sells the marked items; in town it sells at once; nothing marked fails with the reason
- [x] #2 The queue limit (8) and cancel from `queued` hold
- [x] #3 Route test with goldens; old inventory goldens unchanged
- [x] #4 Live checklist added to the doc
<!-- AC:END -->
