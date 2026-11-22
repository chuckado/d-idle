---
id: TASK-823
title: 'uinav2: one shared navigation on every page, Skills area sub-tabs, old layout bar (review E4)'
status: Proposed
result: "Verified offline, not on a real phone: Go tests (TestNavEveryPage, TestNavJS, TestPageLinkMatrix, TestFitNavJudge), node test of nav.js (renders, one aria-current, More sheet focus in, Escape, focus back, Skills sub-tabs), and the fixture-viewer fit check on the worker host (full run: 1665 ok against 1634 on origin/main, no page that passed before fails except one flaky must-fail that passes alone; all nav-* pages and navdefect-* pass or fail as they must). Built: web/nav.json, nav.js, nav.css; every page loads it (live, Gear, Skills, Points, Vault, Characters, Return, Policy, Hall of Fame); the ... menu and the Inventory phone list are the More sheet; Points is the Points and Build sub-tabs of the Skills area; Old layout name and standing Switch back bar; Edit layout hidden on a phone; Hall of Fame in More, Pacts unlinked. Fit findings fixed: phone cockpit frame, Characters and Vault pages share the window with the bar and band. Pre-existing fit reds (vault scroll, 2 px overflows on Gear) are unchanged. Not done: Loot rules from non-live pages, Settings sheet beyond Theme (needs E10)."
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
- [x] #1 TestNavEveryPage, TestNavJS and the link matrix fail on the old pages and pass now
- [x] #2 nav.json/nav.js/nav.css: five destinations, bottom bar under 700 px, band links above, one aria-current, More sheet replaces the ... menu on every page
- [x] #3 Points merged into the Skills area as sub-tabs Tree, Points, Build; old routes keep working
- [x] #4 Old layout naming and standing Switch back bar; Hall of Fame in More, Pacts unlinked
- [x] #5 uishots fit pages nav-<page> at 390x844 and 1366x768 pass and navdefect-* fail; cockpit fit check run, findings fixed or listed
- [x] #6 Before/after pictures in docs/ui/nav-shared; docs/ui-design-system.md section 4.6
<!-- AC:END -->
