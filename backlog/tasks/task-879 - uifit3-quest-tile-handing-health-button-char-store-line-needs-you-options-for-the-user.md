---
id: TASK-879
title: 'uifit3: quest tile handing, health button, char-store line; needs-you options for the user'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T08:54:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 394000
card: "uifit3"
owns:
  - "internal/watch/web/panels/quest-strip/"
  - "internal/watch/web/panels/health/panel.css"
  - "tools/uishots/fit.go"
  - "tools/uishots/fit_char_test.go"
  - "docs/ui/fit-triage-2026-11-23.md"
  - "docs/ui/fit-triage/"
exclusive: []
contracts: []
result: "R4b quest tile, R6 health button fixed and re-measured (quest-* and strip-unknown pass at all sizes on the worker host); R7 line 135 to 140 with rows; R5 needs-you at 360x640 left for the user with two options and pictures in docs/ui/fit-triage-2026-11-23.md. Full fit not re-run after this card."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 quest tile handing and health button fixed, re-measured on the worker host
- [x] #2 char-store line moved with rows at 136, 140, 141, 145
- [x] #3 needs-you options with pictures in the triage doc
<!-- AC:END -->
