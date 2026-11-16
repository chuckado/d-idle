---
id: TASK-561
title: 'pausefix: Pause answered ''could not do that'' although the game froze'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T19:08:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 293000
card: "pausefix"
owns:
  - "internal/watch/pauseapi.go"
  - "internal/watch/pausefix_test.go"
  - "internal/watch/web/pauseui.js"
  - "internal/watch/web/panels/orders/panel.js"
exclusive: []
contracts: []
result: "Fixed: the pause view had no ok flag, so the Orders Pause verb toasted could not do that after a successful freeze. The reply now carries ok (false plus the reason on a refusal); pauseui judges by GET read-back; the Orders panel accepts a pause view. Tests with the real shape. Not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Root cause: POST /api/pause (relayed by the Orders Pause verb through /api/commands) returned the view without the ok flag; the Orders panel read that as a refusal. Not run live.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 the pause reply carries ok:true; a refusal keeps its reason
- [x] #2 UI judges by read-back, never a generic line when the action worked
- [x] #3 tests with the real failure shape (view without ok)
<!-- AC:END -->
