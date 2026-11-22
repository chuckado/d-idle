---
id: TASK-752
title: 'mgwatchdog: merge queue service unit, heartbeat, three alarms, phase budgets, status exit code'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 354000
card: "mgwatchdog"
owns:
  - "docs/process/pg-ops-mgwatchdog.md"
exclusive: []
contracts: []
result: "Merge queue watchdog (ops repo): heartbeat, three alarms (worker dead or no heartbeat 10 min; item running over 45 min; eligible items idle 30 min with CI green), 5-minute user timer that restarts a dead worker (max 3 per hour, not after a deliberate stop), phone push once per alarm and a board tile, phase budgets (preflight 5, gates 30, push 3, ci-wait 12 min) that name the phase of a hang, `mergeq status` line and exit 2. Offline tests (6) pass; timer installed and green on the live queue; heartbeat and phase budgets take effect at the next worker restart (U). Worker service unit exists but is not enabled."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design 4.7, 7 card 14. Private ops repo work. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Service unit and heartbeat
- [x] #2 three alarms and phase budgets
- [x] #3 status exit code
<!-- AC:END -->
