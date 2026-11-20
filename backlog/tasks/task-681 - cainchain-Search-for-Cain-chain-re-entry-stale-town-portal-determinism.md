---
id: TASK-681
title: 'cainchain: Search for Cain chain re-entry, stale town portal, determinism'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T02:38:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 333000
card: "cainchain"
owns:
  - "internal/task/act1.go"
  - "internal/task/cainreenter_test.go"
  - "internal/town/trip.go"
  - "internal/town/portalonly_test.go"
  - "internal/nav/route.go"
  - "docs/probes/provingground-cainchain-2026-11-20.md"
  - "internal/town/town.go"
  - "internal/data/db.go"
  - "internal/bot/questboard.go"
  - "internal/bot/task.go"
  - "internal/bot/runner.go"
  - "internal/bot/staletarget.go"
  - "internal/bot/questboard_goal_test.go"
exclusive: []
contracts: []
result: "cain.chain GREEN on the builder (V, CHEATED setup, mechanics only): Paladin 11 twice (jobs 20261119-232655-7217, 20261120-010053-a896), Paladin 5 (20261120-000256-0f09), Sorceress 3 (20261120-002704-6dae). Fixes: stale town portal refused for portal-only levels, cage re-entry through Stony Field, no leveling plan inside Tristram, nav-fault skips do not stop a started chunk, idle ticks spend time. Not verified live on shipped d2d."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: cain.chain was flaky (stale town portal, re-entry, level planner inside Tristram, frozen clock). Done when: cain.chain green on Paladin 11, 5 and Sorceress 3 and the same-host repeat gives the same verdict.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 cage step re-enters Tristram through the Stony Field portal
- [x] #2 BackToField refuses a stale portal for a portal-only level
- [x] #3 cain.chain green 3 runs Paladin 11, 5, Sorceress 3 and a determinism pair
<!-- AC:END -->
