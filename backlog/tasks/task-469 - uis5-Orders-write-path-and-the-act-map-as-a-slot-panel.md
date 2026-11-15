---
id: TASK-469
title: 'uis5: Orders write path and the act map as a slot panel'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T04:50:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 266000
card: "uis5"
owns:
  - "internal/watch/commands.go"
  - "internal/watch/commands_post.go"
  - "internal/watch/commands_post_test.go"
  - "internal/watch/web/panels/orders/"
  - "internal/watch/web/panels/zone-map/"
  - "internal/watch/testdata/zonemap/"
  - "tools/uishots/"
  - "internal/watch/zonemap_test.go"
  - "internal/watch/testdata/orders_dom.js"
  - "internal/watch/testdata/orders_dom.txt"
  - "internal/watch/mapview.go"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/shell.css"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the Orders panel and the act map were read-only; the player needs two clicks from a zone to an order. Done when: POST /api/commands dispatches over the old routes, the verb row, More sheet, Undo and keys are live, the zone-map slot panel draws the act and its menu opens beside the map (never over a zone button), and the fit check passes at all eight sizes with the menu open.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Answer-route tests for each intent: gates, and the same resulting state as the old route
- [ ] #2 DOM goldens: the grid for the fixture act with here, target, two pins, a failed mark; the menu for field, here, town, hard, locked, charter, loop locked, loop open, queue locked, queue full
- [ ] #3 At every size of the fit rule the map tile shows all ten rows with readable names (no row under 20 px) and the open menu overlaps no zone button (node test of the two rectangles; fit check with the menu open)
- [ ] #4 From a zone click to an accepted order is two clicks (node test)
- [ ] #5 The old orders-map card still places an order in Classic
<!-- AC:END -->
