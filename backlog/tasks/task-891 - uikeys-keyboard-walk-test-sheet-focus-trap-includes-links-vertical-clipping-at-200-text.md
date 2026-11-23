---
id: TASK-891
title: 'uikeys: keyboard walk test, sheet focus trap includes links, vertical clipping at 200% text'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T11:14:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 400000
card: "uikeys"
owns:
  - "internal/watch/testdata/keywalk_test.cjs"
  - "internal/watch/settingsjs_test.go"
  - "internal/watch/web/sheet.js"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "docs/ui-design-system.md"
exclusive: []
contracts: []
result: "Found and fixed: the shared sheet focus trap skipped links (the Loot rules link in Settings). TestKeyboardWalk (stub DOM) walks Settings, More and the bottom bar; TestNoPositiveTabindex; ts200 pages now also judge vertical clipping, must-fail page tsdefect-vcut fails as it must, the four ts200 pages pass on the worker host. Map: the Area map with a trail and an objective (variant walked) already passes the --map check with drawn pixels (390x844, 1366x768); new fit page zone-act-objective expects the Act map to draw its target (data-target) and passes at every size. Not a real-browser keyboard walk."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 keyboard walk test on the Settings and More sheets; no positive tabindex
- [x] #2 ts200 pages fail for vertical clipping; must-fail page tsdefect-vcut
<!-- AC:END -->
