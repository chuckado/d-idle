---
id: TASK-876
title: 'uichars: Characters page shows a hero that is not saved yet (Creating row, orphan folder)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T08:42:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 394000
card: "uichars"
owns:
  - "internal/charsel/store.go"
  - "internal/charsel/manager.go"
  - "internal/charsel/store_unsaved_test.go"
  - "internal/watch/web/characters.js"
  - "internal/watch/web/characters.json"
  - "internal/watch/charactersui_test.go"
  - "tools/uishots/fixtures_chars.go"
  - "docs/charselect.md"
exclusive: []
contracts: []
result: "Characters page: a Creating <name> row with the plain line until the first save lists the hero; a parked folder with no .d2s is returned as unsaved and shown as not saved yet (no Delete: nothing to lose, the typed-name delete rule of saved heroes is unchanged). Node DOM rows, Go test, fit pages chars-unsaved pass at both sizes on the worker host. Not tried against the real game."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Creating row until the first save lists the hero, with the plain line
- [x] #2 a parked folder with no .d2s is shown as not saved yet (never nothing); must-fail DOM rows
- [x] #3 fit pages chars-unsaved at 390x844 and 1366x768 pass
<!-- AC:END -->
