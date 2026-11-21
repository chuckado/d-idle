---
id: TASK-745
title: 'mgonejob: merge gate as one locked builder job; failing tests and tail stored; positive ran-marker; flake ledger'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:49:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 351000
card: "mgonejob"
owns:
  - "docs/process/pg-ops-mgonejob.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design 4.6 and 4.7, 7 card 11. Private ops repo work. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Ops repo: gate is one locked builder job, stores failing tests and tail
- [ ] #2 positive marker that tests ran (a result under 20 s did not run)
- [ ] #3 flake ledger
- [ ] #4 described by role here; commits live in the ops repo
<!-- AC:END -->
