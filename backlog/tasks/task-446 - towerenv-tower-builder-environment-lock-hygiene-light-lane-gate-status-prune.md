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
- [ ] #1 lockrun: lock not inherited by daemons, watchdog, drilled
- [ ] #2 light lane drilled while the main lock is held
- [ ] #3 towergate: whole-repo vet/tests/race/leakcheck/check.sh in one locked job
- [ ] #4 tower-status and prune policy
- [ ] #5 doc with rollback and how agents use the builder
<!-- AC:END -->
