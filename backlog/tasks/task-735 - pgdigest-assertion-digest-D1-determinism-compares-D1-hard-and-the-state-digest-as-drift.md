---
id: TASK-735
title: 'pgdigest: assertion digest D1; determinism compares D1 hard and the state digest as drift'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:45:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 342000
card: "pgdigest"
owns:
  - "internal/scenario/determinism.go"
  - "cmd/d2scenario/cmds.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design section 2.6, 7 card 2 (user Q1: verdict + assertion digest gate; full-state digest is a drift report). Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 D1 = scenario, class, seed and per-assertion facts (discrete by value, ranges by bucket)
- [ ] #2 determinism gates on verdict + D1; whole-state digest reported as drift only
- [ ] #3 must-fail test: gold 700 vs 712 stays equal, quest word differs fails
- [ ] #4 existing scenario verdicts unchanged (listed)
<!-- AC:END -->
