---
id: TASK-424
title: 'uis1: Shell foundations and the default flip'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T15:47:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 255000
card: "uis1"
owns:
  - "internal/watch/web/shell.js"
  - "internal/watch/web/layout.js"
  - "internal/watch/web/shell.json"
  - "internal/watch/web/layout.json"
  - "internal/watch/web/layouts/"
  - "internal/watch/web/tokens.css"
  - "internal/watch/web/shell.css"
  - "internal/watch/layout.go"
  - "internal/watch/layout_test.go"
  - "tools/uilint/"
  - "tools/uishots/layout-rows.json"
  - "tools/uishots/genlayouts.js"
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
- [ ] #1 A test loads `testdata/compat/layout-v1-{mockup,live,command}.json` and the old `layout.json` (now classic) and asserts the same placed ids and cells as before this change
- [ ] #2 Standard and Simple resolved against today's panel list: no unknown id, no overlap, no empty cell; every fallback of D.4 is shown in its cell (node test of the plan)
- [ ] #3 An `optIn` panel is not mounted under a layout that does not name it (node test with a fixture manifest)
- [ ] #4 `?layout=classic` renders the page as before S1 (uishots diff against the present golden, said in the result)
- [ ] #5 The `...` menu switches presets without a POST; Edit layout still saves and Reset returns to the chosen preset
- [ ] #6 uilint knows `status`, `fallback`, `optIn`; token compat test green; pictures at 390 and 1440 looked at for Standard, Simple and Classic
<!-- AC:END -->
