---
id: TASK-877
title: 'uiloot: Loot rules reachable from Settings > Rules on every page'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T08:45:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 394000
card: "uiloot"
owns:
  - "internal/watch/web/settings.js"
  - "internal/watch/web/shell.js"
  - "internal/watch/testdata/settings_test.cjs"
  - "internal/watch/settingsjs_test.go"
  - "tools/uishots.sh"
exclusive: []
contracts: []
result: "Settings > Rules has a Loot rules link (the old Inventory entry) to /watch?live=1&lootrules=1; shell.js opens the sheet from that param. DOM rows, Go test, fit pages lootrules-link and settings-* pass on the worker host. Not tried against a real hero."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Settings > Rules has a Loot rules link; the live screen opens the sheet from ?lootrules=1
- [x] #2 fit page lootrules-link passes at all sizes; DOM and Go rows pass
<!-- AC:END -->
