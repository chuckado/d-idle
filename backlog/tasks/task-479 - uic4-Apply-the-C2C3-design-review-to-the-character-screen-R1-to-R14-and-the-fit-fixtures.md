---
id: TASK-479
title: 'uic4: Apply the C2C3 design review to the character screen (R1 to R14) and the fit fixtures'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T06:38:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: high
ordinal: 268000
card: "uic4"
owns:
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/web/inventory-tip.css"
  - "internal/watch/web/inventory-tip.js"
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-actions.css"
  - "internal/watch/web/inventory-drag.js"
  - "internal/watch/web/inventory-merc.js"
  - "internal/watch/web/inventory-merc.css"
  - "internal/watch/web/inventory.js"
  - "internal/watch/web/tokens.css"
  - "internal/watch/invui.go"
  - "internal/watch/tipcontrast_test.go"
  - "internal/watch/uic4_test.go"
  - "internal/watch/testdata/uic4/"
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
- [ ] #1 R1 to R13 of docs/ui/redesign-2026-11-14/reviews/c2c3-review.md applied
- [ ] #2 R14 fit fixtures: no mercenary, dead merc, phone storage tab, long names, pending state, light tooltip fail the fit check or tests when the defect returns
- [ ] #3 tools/uishots.sh --fit passes at the eight sizes and the renders were looked at
- [ ] #4 gate (labrun tower --gate) passes
<!-- AC:END -->
