---
id: TASK-466
title: 'uib4: Until-level for zone orders'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T04:16:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 265000
card: "uib4"
owns:
  - "internal/watch/orders.go"
  - "internal/watch/orders_untillevel_test.go"
  - "internal/watch/hub.go"
  - "internal/watch/http_answer.go"
  - "internal/watch/funwire.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

POST /api/orders accepts until: level with level: N (go and fight). Order gains untilLevel (additive, omitempty); Hub.TickOrders (called by Fun.Settle) ends it done at the level; the queue's own entries are untouched.

result: built and tested offline (sim hub, funRig Settle, Sorceress and Paladin); not verified live.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Table tests for validate; an order ends at the level (hub test with the sim); the /api/orders reply keeps every old field (golden)
- [x] #2 The queue's until-level behaviour is unchanged
<!-- AC:END -->
