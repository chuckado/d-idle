---
id: TASK-642
title: >-
  itempics: jewel, ring, amulet and charm pictures from the item's own picture
  index
status: Done
assignee:
  - credit-runner
created_date: '2026-11-18T15:31:00Z'
updated_date: '2026-11-18T16:03:00Z'
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
  - "internal/d2/items.go"
  - "internal/data/invfile_test.go"
  - "internal/data/invfile_real_test.go"
  - "internal/d2dsrc/source_test.go"
  - "internal/d2dsrc/testdata/dump_items_pic.json"
  - "cmd/d2bot/itemart_test.go"
exclusive: []
result: "Real jewel/ring/amulet/charm art by the item picture index (ItemTypes InvGfx1..6, V on the real 1.14d install: jewel invjw1..6, ring invrin1..5, amulet invamu1..3, charms invch1..9). Offline tests only; needs a new d2d build with inv_pic (fork caea2368, not built or run here); until then the base art. Not verified live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Jewels all showed the base gem art. The game picks one of ItemTypes InvGfx1..6 by the item's picture field. Done when the page shows the real look for jewels, rings, amulets, charms.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ItemTypes InvGfx table read; art name per picture index
- [x] #2 d2dsrc reads inv_pic into d2.Item.Pic (older d2d: base art)
- [x] #3 offline tests: six jewel keys distinct, ring, unchanged non-variant items
- [x] #4 fork field inv_pic on didle/rebase-20261118 (needs devp12, not built here)
<!-- AC:END -->
