---
id: TASK-933
title: 'checkpoint1: checkpoint portals phase 1 (cast, retreat, return after death)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-25T17:03:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 435000
card: "checkpoint1"
owns:
  - "internal/bot/checkpoint.go"
  - "internal/bot/checkpoint_test.go"
  - "internal/plandata/checkpoint_areas.go"
  - "internal/env/tuning_bot.go"
  - "docs/knobs.md"
  - "docs/design/danger-protocol-2026-11-25.md"
  - "internal/bot/loop.go"
  - "internal/bot/goal.go"
  - "internal/bot/observe.go"
  - "internal/bot/runner.go"
  - "internal/bot/state.go"
exclusive: []
contracts: []
result: "phase 1 built behind bot.checkpointEnable (default off): record, proactive cast, retreat, death route, arrival check; offline tests on two classes pass; not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Checkpoint record and state machine behind bot.checkpointEnable (default off)
- [x] #2 Proactive cast at a danger area, retreat at low life, return after death with the arrival check
- [x] #3 Tests on two classes; knobs docs regenerated; design doc updated with built vs deferred
<!-- AC:END -->
