---
id: TASK-647
title: 'unidpopup: unidentified card sheet must not auto-open'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T16:59:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 323000
card: "unidpopup"
owns:
  - "internal/watch/web/unlockui.js"
  - "internal/watch/unlockui_quiet_test.go"
  - "docs/unlock-idreveal.md"
exclusive: []
contracts: []
result: "Cause: the Cards sheet (Unidentified item) auto-opened for every new face-down card and again after each page load. Now never auto-opens; the Cards tab shows (N). Offline node test only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Face-down cards never open the sheet by themselves; one count on the Cards tab
- [x] #2 Test: 50 unflipped cards open the sheet 0 times
<!-- AC:END -->
