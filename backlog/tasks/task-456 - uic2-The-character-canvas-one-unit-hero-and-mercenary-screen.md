---
id: TASK-456
title: 'uic2: The character canvas (one-unit hero and mercenary screen)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T22:53:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "uic2"
owns:
  - "internal/watch/web/inventory.js"
  - "internal/watch/web/inventory.css"
  - "internal/watch/web/inventory-merc.js"
  - "internal/watch/web/inventory-merc.css"
  - "internal/watch/web/inventory-sets.js"
  - "internal/watch/web/inventory-sets.css"
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/invui.go"
  - "internal/watch/invapi.go"
  - "internal/watch/charscreen_test.go"
  - "internal/watch/testdata/charscreen/"
  - "tools/uishots.sh"
  - "tools/uishots/pages.go"
  - "tools/uishots/fixtures.go"
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
- [ ] #1 Fit check passes for the character page at all 8 sizes (1440x900, 1920x1080, 1366x768, 1914x930, 1904x1000, 1536x730, 1280x620, 390x844)
- [ ] #2 DOM goldens: hero only; each mercenary state (none, dead, desert, ironwolf, unknown, absent); the requirement marks over C1's four views and one cell per reason code
- [ ] #3 No rule in the page (a test greps the scripts for requirement arithmetic and class lists)
- [ ] #4 Every action of the old page still works (old inventory, merc, sets and actions tests pass unchanged; the cards page stays at ?view=classic)
- [ ] #5 Renders at 1914x930, 1536x730 and 1280x620 looked at, critique written in the result
<!-- AC:END -->
