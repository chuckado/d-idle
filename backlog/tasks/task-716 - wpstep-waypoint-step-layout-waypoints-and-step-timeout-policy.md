---
id: TASK-716
title: 'wpstep: waypoint step layout waypoints and step timeout policy'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:38:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 341000
card: "wpstep"
owns:
  - "internal/task/executor.go"
  - "internal/task/timeoutpolicy_test.go"
  - "internal/d2map/static.go"
  - "internal/d2map/tracker.go"
  - "internal/d2map/tracker_test.go"
  - "internal/d2dworld/"
  - "internal/bot/task.go"
  - "tools/provingground/scenarios/cain.json"
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
- [ ] #1 RED test: layout waypoint of unknown class is found
- [ ] #2 step timeout policy: one extension on progress, then park with reason
- [ ] #3 scenario act1.wp.dark-wood green on the builder (layoutRead 0 and 1)
<!-- AC:END -->
