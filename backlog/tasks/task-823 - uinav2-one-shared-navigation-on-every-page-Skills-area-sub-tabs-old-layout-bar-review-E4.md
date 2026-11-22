---
id: TASK-823
title: 'uinav2: one shared navigation on every page, Skills area sub-tabs, old layout bar (review E4)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T20:46:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 375000
card: "uinav2"
owns:
  - "internal/watch/web/nav.json"
  - "internal/watch/web/nav.js"
  - "internal/watch/web/nav.css"
  - "internal/watch/linkmatrix_test.go"
  - "internal/watch/vaultlink_test.go"
  - "internal/watch/testdata/nav_test.cjs"
  - "internal/watch/testdata/nav_dom.cjs"
  - "internal/watch/invui.go"
  - "internal/watch/skillsui.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/returnscreen.go"
  - "internal/watch/policysheet.go"
  - "internal/watch/hall.go"
  - "internal/watch/http_view.go"
  - "internal/watch/web/watch.html"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/components.css"
  - "internal/watch/web/utilities.css"
  - "internal/watch/web/app.js"
  - "internal/watch/web/inventory-canvas.css"
  - "internal/watch/web/inventory-canvas.js"
  - "internal/watch/web/vault.html"
  - "internal/watch/web/vault.js"
  - "internal/watch/web/vault.json"
  - "internal/watch/web/vault.css"
  - "internal/watch/web/characters.html"
  - "internal/watch/web/characters.js"
  - "internal/watch/web/characters.json"
  - "internal/watch/web/characters.css"
  - "internal/watch/web/points.css"
  - "tools/uishots.sh"
  - "tools/uishots"
  - "docs/ui/nav-shared"
  - "docs/ui-design-system.md"
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
- [ ] #1 TestNavEveryPage, TestNavJS and the link matrix fail on the old pages and pass now
- [ ] #2 nav.json/nav.js/nav.css: five destinations, bottom bar under 700 px, band links above, one aria-current, More sheet replaces the ... menu on every page
- [ ] #3 Points merged into the Skills area as sub-tabs Tree, Points, Build; old routes keep working
- [ ] #4 Old layout naming and standing Switch back bar; Hall of Fame in More, Pacts unlinked
- [ ] #5 uishots fit pages nav-<page> at 390x844 and 1366x768 pass and navdefect-* fail; cockpit fit check run, findings fixed or listed
- [ ] #6 Before/after pictures in docs/ui/nav-shared; docs/ui-design-system.md section 4.6
<!-- AC:END -->
