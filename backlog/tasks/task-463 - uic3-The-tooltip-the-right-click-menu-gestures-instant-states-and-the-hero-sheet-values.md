---
id: TASK-463
title: 'uic3: The tooltip, the right-click menu, gestures, instant states and the hero sheet values'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T04:08:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 265000
card: "uic3"
owns:
  - "internal/watch/web/inventory-tip.js"
  - "internal/watch/web/inventory-tip.css"
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-actions.css"
  - "internal/watch/web/inventory-drag.js"
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/testdata/uic3/"
  - "internal/watch/herosheet_test.go"
  - "internal/watch/hero.go"
  - "internal/bot/heroattrs.go"
  - "internal/bot/heroattrs_test.go"
  - "cmd/d2bot/readywire.go"
  - "tools/uishots/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 DOM goldens: tooltip (no requirements, each unmet kind, class item, mercenary block, comparison with deltas, both rings), menu per place, each feedback state
- [ ] #2 Tooltip never leaves the viewport nor overlaps its cell at the eight sizes (node test of placement)
- [ ] #3 A refused requirement or action never renders as done (node test)
- [ ] #4 Hero Str Dex Vit Energy defense damage in /api/hero stats, absent when unread, tiles bound
- [ ] #5 Fit check passes at all 8 sizes incl. tooltip open and menu open fixtures; renders looked at
<!-- AC:END -->
