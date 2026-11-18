---
id: TASK-637
title: 'itemcolour2: item tint parse, art remap and UI key'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T14:46:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 320000
card: "itemcolour2"
owns:
  - "internal/art/"
  - "internal/d2dsrc/"
  - "internal/watch/web/inventory.js"
  - "internal/watch/artembed_test.go"
  - "internal/status/inventory.go"
  - "internal/d2/items.go"
  - "cmd/d2bot/display.go"
  - "cmd/d2bot/itemart_test.go"
exclusive: []
contracts: []
result: "Offline only, not verified live. Bot parses inv_tint/inv_colour (field shape from the design, U until devp11), art key item-<invfile>~tTcC remaps through items/palette/<set>.dat from the install at runtime, plain fallback; inventory page appends the tint. Merc items not yet tinted."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 d2dsrc parses inv_tint/inv_colour into d2.Item.Tint, graceful when absent
- [x] #2 internal/art serves item-<invfile>~tTcC through the install colormap, plain fallback
- [x] #3 inventory page appends the tint to the art key, no CSS colour
<!-- AC:END -->
