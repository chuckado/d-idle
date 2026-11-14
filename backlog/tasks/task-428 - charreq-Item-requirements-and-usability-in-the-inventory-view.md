---
id: TASK-428
title: 'charreq: Item requirements and usability in the inventory view'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T15:53:00Z'
updated_date: '2026-11-14T15:53:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 255000
card: "charreq"
owns:
  - "internal/invdesk/usable.go"
  - "internal/invdesk/usable_test.go"
  - "internal/invdesk/desk.go"
  - "internal/invdesk/sets.go"
  - "internal/itemreq/"
  - "internal/leveler/score.go"
  - "internal/leveler/itemreasons.go"
  - "internal/leveler/judge.go"
  - "internal/leveler/judge_test.go"
  - "internal/leveler/playerequip.go"
  - "internal/watch/testdata/itemusable/"
  - "docs/inventory-ui.md"
exclusive: []
contracts: []
result: "req and usable (hero, merc) on the item view, one rule via leveler.Judge; invdesk and watch tests pass on herdr; the 21-case table test in internal/leveler/judge_test.go passed on the worker host before the stop order, whole-repo and race gates not run; not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the page must show what an item requires and who can use it, from the equip path's own rule. Done when: req and usable are on the item view (schema 8).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The 21-row table of I.10 as a table test bound to real rows of the game tables (run with D2_MPQ_DIR), at least Sorceress, Paladin, Amazon, Barbarian and the four mercenary types; rows the tables cannot settle are listed, not guessed
- [x] #2 A property test: for every pack item, usable.hero.ok is true exactly when canWear is non-empty for a wearable kind (one rule, two views)
- [x] #3 View goldens for both classes plus a mercenary fixture; inventory compat goldens unchanged
- [x] #4 Set apply test: an item under its requirement is skipped with the reason and nothing replaces it
<!-- AC:END -->
