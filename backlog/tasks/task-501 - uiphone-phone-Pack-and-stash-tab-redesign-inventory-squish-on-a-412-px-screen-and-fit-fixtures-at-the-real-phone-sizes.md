---
id: TASK-501
title: 'uiphone: phone Pack and stash tab redesign (inventory squish on a 412 px screen) and fit fixtures at the real phone sizes'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T15:00:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: high
ordinal: 276000
card: "uiphone"
owns:
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/inventory.js"
  - "internal/watch/invui.go"
  - "internal/watch/testdata/charscreen/"
  - "tools/uishots/"
  - "tools/uishots.sh"
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
- [ ] #1 Phone storage tab: grids use the full width (cell at least 32 px), art cells show only the picture, belt row and stash full width, selected item as a bottom sheet with 44 px buttons, one chip row with Repair and Sell, folded town errands, the lock line wraps, header keeps a pages menu
- [ ] #2 Fit fixtures at 412x780, 393x700, 360x640, 412x890 and 390x844 with real-looking data (art, full pack with 2x3 and 2x4 items, seven-potion belt, long name, sheet, errands); must-fail pages for label over art, clipped button row, horizontal overflow, truncated lock line, cut chips, small cells
- [ ] #3 tools/uishots.sh --fit passes at every size and the renders at 412x780, 360x640 and 390x844 (dark and light) were looked at
- [ ] #4 the whole-repo gate passes on the builder
<!-- AC:END -->
