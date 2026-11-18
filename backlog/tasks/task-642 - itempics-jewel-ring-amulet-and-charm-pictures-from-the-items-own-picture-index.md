---
id: TASK-642
title: 'itempics: jewel, ring, amulet and charm pictures from the item''s own picture index'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T15:31:00Z'
labels:
  - d2d
  - feature
dependencies: []
priority: medium
ordinal: 321000
card: "itempics"
owns:
  - "internal/data/invfile.go"
  - "internal/data/invfiles.go"
  - "internal/data/items.go"
  - "internal/data/schema.go"
  - "cmd/d2bot/display.go"
  - "internal/d2dsrc/items.go"
  - "internal/bot/catalog.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Jewels all showed the base gem art. The game picks one of ItemTypes InvGfx1..6 by the item's picture field. Done when the page shows the real look for jewels, rings, amulets, charms.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 ItemTypes InvGfx table read; art name per picture index
- [ ] #2 d2dsrc reads inv_pic into d2.Item.Pic (older d2d: base art)
- [ ] #3 offline tests: six jewel keys distinct, ring, unchanged non-variant items
- [ ] #4 fork field inv_pic on didle/rebase-20261118 (needs devp12, not built here)
<!-- AC:END -->
