---
id: TASK-425
title: 'uib7: The Loop order and its unlock'
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
card: "uib7"
owns:
  - "internal/watch/orderloop.go"
  - "internal/watch/orderloop_test.go"
  - "internal/watch/testdata/orderloop/"
  - "data/unlocks/ladder.json"
  - "internal/watch/orderqueue.json"
  - "cmd/d2bot/loopwire.go"
  - "docs/loop-order.md"
  - "cmd/d2bot/loopwire_test.go"
exclusive: []
contracts: []
result: "Built offline, not verified live: OrderLoop controller + GET/POST /api/orderloop, ladder row orderloop (event manual_repair_cycle), orderqueue now opens on loop_rounds_3 (Den no longer opens it; open rows stay open), tests for Sorceress and Paladin, docs/loop-order.md with the live checklist."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the user's progression (single orders, then the Loop, then the queue). Done when: the loop controller, its API, the ladder rows and docs/loop-order.md are in and tested offline. Small additive hooks outside the owned list: internal/watch/funwire.go, cmd/d2bot/funwire.go and repairwire.go (one line each), and the Den-trigger tests of the queue (internal/watch/unlockorderqueue_test.go, cmd/d2bot/unlockwire_test.go) moved to the new trigger.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Sim tests for two classes: three rounds run go -> repair -> go; damaged skips the smith when nothing is damaged; each stop condition ends the loop with its reason; restart resumes the round
- [x] #2 Ladder tests: orderloop opens on the event; orderqueue opens after three rounds; a hero with orderqueue already open keeps it
- [x] #3 Route tests with goldens (testdata/orderloop/view-*.json), 409 while locked
- [x] #4 docs/loop-order.md with the tiers, the spec and a live checklist
<!-- AC:END -->
