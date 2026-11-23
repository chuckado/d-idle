---
id: TASK-860
title: 'uiframe: Return, Policy and Hall on the shared frame, no app.css or old .btn'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:47:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 389000
card: "uiframe"
owns:
  - "internal/watch/returnscreen.go"
  - "internal/watch/policysheet.go"
  - "internal/watch/frame_test.go"
  - "internal/watch/returnscreen_test.go"
  - "internal/watch/testdata/returnscreen/"
  - "internal/watch/testdata/policysheet/"
exclusive: []
contracts: []
result: "Return and Policy pages moved to the shared vocabulary (d-btn/d-card/d-hint/d-badge/d-empty, --d-* tokens), app.css no longer linked by them (it can go with E11 once nothing else links it: check gate.go); Hall already clean. fit pages pass on the worker host at both sizes; not looked at live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Return and Policy load no app.css and use d-btn, d-card, d-hint, d-badge, d-empty and --d-* tokens; Hall checked
- [x] #2 must-fail Go test TestReturnPolicyHallOnSharedFrame; fit pages nav-return, nav-policy, nav-hall pass at 390x844 and 1366x768
<!-- AC:END -->
