---
id: TASK-438
title: 'uis1b: The cockpit template in the shell, and the fit check'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T20:52:00Z'
updated_date: '2026-11-14T20:52:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 259000
card: "uis1b"
owns:
  - "internal/watch/web/shell.js"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/layout.js"
  - "internal/watch/web/layoutedit.js"
  - "internal/watch/web/shell.json"
  - "internal/watch/web/layouts/cockpit.json"
  - "internal/watch/layout.go"
  - "internal/watch/layout_test.go"
  - "internal/watch/panels.go"
  - "internal/watch/http_view.go"
  - "tools/uilint/"
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "internal/watch/testdata/layout_cells.js"
  - "internal/watch/testdata/layout_cells.golden"
  - "internal/watch/testdata/layout_cockpit.js"
  - "internal/watch/testdata/layout_edit.js"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the first attempt's free grid cannot fit a screen (content-height rows). The shell gains a second, additive kind of layout: template cockpit/1 with named slots (band, stage, strip, rail, railWide, phoneTabs) that fills the viewport, next to the unchanged free grid; tools/uishots gets the fit check that proves a page fits. Done when: the criteria below are ticked and the result says what was verified.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every existing layout (stored, Classic, Standard, Simple, Stream, Debug, the three old presets) renders the same cells as before (the S1 compat test still passes, plus a new one for template-less files)
- [ ] #2 ?layout=cockpit with the panels that exist today (video, hero-meters) fills the viewport and does not scroll at the sizes of the fit rule; empty slots are left out without holes (fit check output in the result)
- [ ] #3 Editor on the cockpit: swap two slots, move both dividers, save, reload: the same arrangement (node test of the plan; pictures looked at)
- [ ] #4 uilint accepts the new keys and rejects an unknown slot name or a non-slot panel in a slot
- [ ] #5 The fit check fails on a deliberately overflowing fixture page and passes on the cockpit
<!-- AC:END -->
