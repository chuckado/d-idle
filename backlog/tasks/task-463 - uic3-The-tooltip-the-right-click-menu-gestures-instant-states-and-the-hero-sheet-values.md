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
  - "tools/uishots.sh"
  - "internal/watch/uic3_test.go"
  - "internal/watch/web/inventory-sets.css"
  - "docs/inventory-ui.md"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the character canvas (uic2) still opened the old click popover, had no real tooltip with requirements and comparison, and its hero stat tiles were blank. Done when: a floating tooltip with signed deltas, a right-click menu with keys and gestures, instant-action states and the missing hero values work and pass the fit check at the eight sizes with the tooltip and the menu open. Brief: docs/ui/redesign-2026-11-14/briefs/C3-tooltip-menu-gestures.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 DOM goldens: tooltip (no requirements, each unmet kind, class item, mercenary block, comparison with deltas, both rings), menu per place, each feedback state
- [ ] #2 Tooltip never leaves the viewport nor overlaps its cell at the eight sizes (node test of placement)
- [ ] #3 A refused requirement or action never renders as done (node test)
- [ ] #4 Hero Str Dex Vit Energy defense damage in /api/hero stats, absent when unread, tiles bound
- [ ] #5 Fit check passes at all 8 sizes incl. tooltip open and menu open fixtures; renders looked at
<!-- AC:END -->
