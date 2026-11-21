---
id: TASK-754
title: 'pgnightly: one nightly scheduler job: T2, T3, pairs, one T4; smoke stage after each push'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:51:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 355000
card: "pgnightly"
owns:
  - "docs/process/pg-ops-pgnightly.md"
  - "tools/provingground/nightly.sh"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design 4.6, 7 card 15. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Nightly job runs T2, T3, pairs and one T4 in the standing worker-host window 23:00-07:00 EST only, at most 3 runs, niced (Q4)
- [ ] #2 smoke set report-only for 7 clean days, then the user decides on blocking (Q2)
<!-- AC:END -->
