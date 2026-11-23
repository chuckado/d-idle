---
id: TASK-867
title: 'uifit1: fit suite triage and fixes (stale rules and phone/desk layout defects)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T07:10:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 390000
card: "uifit1"
owns:
  - "tools/uishots/"
  - "tools/uishots.sh"
  - "docs/ui/fit-triage-2026-11-23.md"
  - "internal/watch/web/utilities.css"
  - "internal/watch/web/shell.css"
  - "internal/watch/web/panels/statusline/panel.css"
  - "internal/watch/web/panels/quest-strip/panel.css"
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
- [ ] #1 triage doc groups every red fit run by rule with a verdict (real, stale, fixture)
- [ ] #2 stale rules relaxed with must-pass and must-fail unit rows (fit_phone_test.go)
- [ ] #3 real defects fixed and re-measured on the worker host; final red count per size in the doc
<!-- AC:END -->
