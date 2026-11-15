---
id: TASK-461
title: 'toastfix: stack toast text unreadable in dark theme'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T01:25:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 264000
card: "toastfix"
owns:
  - "internal/watch/web/app.css"
  - "internal/watch/toastcontrast_test.go"
exclusive: []
contracts: []
result: "Fixed: .toast stack rule now sets color var(--text); inverted rule scoped to #toast. Contrast test dark+light. Not verified in a browser render."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Stack toast sets color var(--text) on --surface; test asserts 4.5:1 in both themes
<!-- AC:END -->
