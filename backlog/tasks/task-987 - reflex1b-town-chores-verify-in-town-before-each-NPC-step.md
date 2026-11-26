---
id: TASK-987
title: 'reflex1b: town chores verify in town before each NPC step'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T03:58:00Z'
updated_date: '2026-11-26T05:13:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 456000
card: "reflex1b"
owns:
  - "internal/town/npc.go"
  - "internal/town/chores.go"
  - "internal/town/intown_test.go"
  - "internal/bot/townguard.go"
  - "internal/bot/townguard_test.go"
  - "internal/bot/goal.go"
  - "internal/bot/loop.go"
  - "internal/bot/runner.go"
  - "internal/bot/state.go"
  - "docs/safety-latency.md"
exclusive: []
result: "Town chores verify in town before every chore and before every NPC walk (town.ErrNotInTown, rest of the chores skipped); the guard aborts a running chore after 2 s outside a town (errLeftTown, no failure count) and the trip restarts from the way home. Regression from the 02:51 death on two classes at bot level (fakes); the town-level test needs game tables and skips on the dev boxes. Offline only, not verified live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Fable review P3 (d). The 2026-11-26 02:51 death: chores ran in Cold Plains.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 town chores check in town before each NPC step and abort the trip otherwise
- [x] #2 regression: trip in chores stage, hero in a field with hostiles (02:51 death), on two classes
<!-- AC:END -->
