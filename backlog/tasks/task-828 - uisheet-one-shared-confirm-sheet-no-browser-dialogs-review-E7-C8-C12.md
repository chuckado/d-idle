---
id: TASK-828
title: 'uisheet: one shared confirm sheet, no browser dialogs (review E7, C8, C12)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T00:16:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 378000
result: "Verified offline, not live: node TestSheetDOM, TestNoBrowserDialogs (grep), TestBuildCardsDOM, go test ./internal/watch; fit pages builds-* at 390x844 and 1366x768 on the worker host. dSheet used by app.js, skills.js, update.js, points.js. Vault and order sheets stay (docs/ui-design-system.md 4.4a)."
card: "uisheet"
owns:
  - "internal/watch/web/sheet.js"
  - "internal/watch/web/buildcards.js"
  - "internal/watch/web/app.js"
  - "internal/watch/web/skills.js"
  - "internal/watch/web/update.js"
  - "internal/watch/web/watch.html"
  - "internal/watch/web/components.css"
  - "internal/watch/web/points.css"
  - "internal/watch/web/shell.css"
  - "internal/watch/pointsui.go"
  - "internal/watch/skillsui.go"
  - "internal/watch/update.go"
  - "internal/watch/http_view.go"
  - "internal/watch/sheet_noprompt_test.go"
  - "internal/watch/buildcards_test.go"
  - "internal/watch/testdata/sheet_test.cjs"
  - "internal/watch/testdata/buildcards_test.cjs"
  - "docs/ui/uisheet"
  - "docs/ui-design-system.md"
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
- [x] #1 web/sheet.js is the one confirm sheet (roles, focus in, Escape, backdrop, focus return) with a node test
- [x] #2 no window.confirm, alert or prompt in internal/watch/web (grep test)
- [x] #3 points, skills, update and app callers use the sheet; fit pages builds-* pass at 390x844 and 1366x768
<!-- AC:END -->
