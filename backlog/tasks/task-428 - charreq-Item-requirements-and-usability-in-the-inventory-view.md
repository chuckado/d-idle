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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The 21-row table of I.10 as a table test bound to real rows of the game tables (run with D2_MPQ_DIR), at least Sorceress, Paladin, Amazon, Barbarian and the four mercenary types; rows the tables cannot settle are listed, not guessed
- [ ] #2 A property test: for every pack item, usable.hero.ok is true exactly when canWear is non-empty for a wearable kind (one rule, two views)
- [ ] #3 View goldens for both classes plus a mercenary fixture; inventory compat goldens unchanged
- [ ] #4 Set apply test: an item under its requirement is skipped with the reason and nothing replaces it
<!-- AC:END -->
