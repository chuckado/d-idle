---
id: TASK-825
title: 'buildcards: Build tab as cards from the catalogue (review E6)'
status: Proposed
result: "Verified offline, not live: Go tests (cards for Paladin, Sorceress, Necromancer, unknown class, detail 404), node DOM golden and sheet tests, fit pages builds-list and builds-detail at 390x844 and 1366x768 on the worker host, pictures in docs/ui/build-cards. Build tab = cards from the catalogue with variants behind them, one confirm sheet, Choose disabled with Auto-allocation is off when the knob is off (default unchanged). The sheet has no picture. Build-specific data lives in plandata build-catalog.json."
assignee:
  - credit-runner
created_date: '2026-11-22T23:48:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 377000
card: "buildcards"
owns:
  - "internal/points/buildcards.go"
  - "internal/points/buildcards_test.go"
  - "internal/points/autoalloc.go"
  - "internal/plandata/catalog.go"
  - "internal/watch/pointsapi.go"
  - "internal/watch/pointsapi_test.go"
  - "internal/watch/pointsui.go"
  - "internal/watch/buildcards_test.go"
  - "internal/watch/web/buildcards.js"
  - "internal/watch/web/points.js"
  - "internal/watch/web/points.css"
  - "internal/watch/testdata/buildcards_test.cjs"
  - "internal/watch/testdata/buildcards_dom.txt"
  - "tools/uishots/fixtures_builds.go"
  - "tools/uishots/fit.go"
  - "tools/uishots/fit.js"
  - "tools/uishots/pages.go"
  - "tools/uishots/uishots_test.go"
  - "tools/uishots.sh"
  - "docs/ui/build-cards"
  - "docs/builds-catalog.md"
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
- [x] #1 Build cards from the catalogue in the build API (cards[], detail on demand), Go tests for Paladin, Sorceress, Necromancer
- [x] #2 Build tab as cards, detail view, one confirm sheet, autoAlloc off = read-only with the visible reason
- [x] #3 DOM golden, sheet focus/Escape node test, fit pages builds-list and builds-detail at 390x844 and 1366x768
- [x] #4 Pictures before/after in docs/ui/build-cards, docs
<!-- AC:END -->
