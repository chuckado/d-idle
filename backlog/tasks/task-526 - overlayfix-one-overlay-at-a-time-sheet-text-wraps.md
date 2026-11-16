---
id: TASK-526
title: 'overlayfix: one overlay at a time, sheet text wraps'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:22:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 283000
card: "overlayfix"
owns:
  - "internal/watch/web/overlay.js"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/tokens.css"
  - "internal/watch/web/panels/zone-map"
  - "internal/watch/testdata/overlay_dom.mjs"
  - "internal/watch/resetui_test.go"
  - "internal/watch/http_view.go"
  - "tools/uishots.sh"
  - "tools/uishots"
  - "docs/ui-overlays.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: user bug report 2026-11-15, a zone sheet painted over the "..." menu and its charter line was cut at the right edge. Cause and policy: docs/ui-overlays.md. Done when: the menu is above every sheet, one overlay is open at a time, sheet text wraps, and the fit suite proves it with must-fail defects.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 One overlay policy: opening the menu closes any sheet, drawer or popover and the reverse; Escape closes the top-most and returns focus
- [ ] #2 The menu is a top-layer popover and is never covered; sheet text wraps, scrolls inside, Pay and go stays visible
- [ ] #3 uishots fit checks and must-fail defects for menu-over-sheet, clipped sheet text, sheet under the header; node DOM test for the overlay manager
- [ ] #4 Renders viewed at 1914x930, 1366x768, 1280x620, 412x780 dark and light
<!-- AC:END -->
