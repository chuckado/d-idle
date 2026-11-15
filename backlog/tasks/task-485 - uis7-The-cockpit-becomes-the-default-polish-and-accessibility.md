---
id: TASK-485
title: 'uis7: The cockpit becomes the default; polish and accessibility'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T08:45:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 270000
card: "uis7"
owns:
  - "internal/watch/web/layout.json"
  - "internal/watch/web/layouts/"
  - "internal/watch/web/components.css"
  - "internal/watch/web/panels/{statusline,needs-count,needs-you,orders,zone-map,area-map,progress,journey-strip,health,log}/panel.css"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/layout.js"
  - "internal/watch/layout_test.go"
  - "internal/watch/testdata/layout_default.js"
  - "internal/watch/testdata/layout_cells.golden"
  - "tools/uishots/layout-rows.json"
  - "tools/uishots/genlayouts.js"
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
- [ ] #1 Fit check passes for the default page (no layout param) at the eight sizes, dark and light, with zero, two and five needs rows and with one and eight orders
- [ ] #2 A stored layout is untouched (test); an auto-pinned Simple follows the default; the Layout menu has Cockpit and returns to every older layout
- [ ] #3 Renders at 1440x900, 1920x1080, 1366x768, 1280x620 and the phone looked at, dark and light, empty and full data
- [ ] #4 Tab order follows reading order; every control reachable by keyboard; contrast 4.5:1; reduced motion
<!-- AC:END -->
