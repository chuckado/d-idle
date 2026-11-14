---
id: TASK-446
title: 'towerenv: tower builder environment (lock hygiene, light lane, gate, status, prune)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T21:46:00Z'
labels:
  - ops
  - chore
dependencies: []
priority: medium
ordinal: 259000
card: "towerenv"
owns:
  - "docs/process/build-hosts-env-2026-11-14.md"
result: "Applied on the builder with no container restart: /data/bin/lockrun (flock -o, watchdog, light lane), towerctl (status, watchdog, prune), towergate; labrun --light/--gate/tower-status/tower-install. Drilled live on the builder with private locks (daemon no longer holds the slot, watchdog frees stuck holders, light lane runs under a held main lock). First real labrun tower --gate: TOWERGATE PASS (check.sh 24s, tests 159s, race 36s, leakcheck 0 hits). Unverified: User Scripts prune schedule (needs NAS root). Doc: docs/process/build-hosts-env-2026-11-14.md."
exclusive:
  - "fork-branch:none"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User go 2026-11-14. Done when the doc exists and the drills are recorded.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 lockrun: lock not inherited by daemons, watchdog, drilled
- [x] #2 light lane drilled while the main lock is held
- [x] #3 towergate: whole-repo vet/tests/race/leakcheck/check.sh in one locked job
- [x] #4 tower-status and prune policy
- [x] #5 doc with rollback and how agents use the builder
<!-- AC:END -->
