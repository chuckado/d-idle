---
id: TASK-955
title: 'invmon: invariant monitor (log only)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:35:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 454000
card: "invmon"
owns:
  - "internal/bot/invariants.go"
  - "internal/bot/invariants_test.go"
  - "docs/invariants.md"
  - "internal/bot/runner.go"
  - "internal/bot/loop.go"
  - "internal/bot/observe.go"
  - "internal/bot/task.go"
  - "internal/env/tuning_bot.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Size: S. P1 of the 2026-11-25 logic review. Logging only.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 table of 13 named invariants checked once per tick, one rate-limited line each, counts in the minute line
- [x] #2 must-fail tests on two heroes (different life, drink and chicken lines, kit lists)
- [x] #3 knob bot.invariants default 1 (log only), 0 off; Deps.OnInvariant oracle hook
<!-- AC:END -->
