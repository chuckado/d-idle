---
id: TASK-733
title: 'pgprov: result records bot commit, d2d hash, host role, load, pace; a run cut by the wall limit is BLOCKED'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:44:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 341000
card: "pgprov"
owns:
  - "internal/scenario/result.go"
  - "cmd/d2scenario/run.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Design section 2.2 H5 and H6, 7 card 1. Done when a pair records its provenance and a cut run cannot pass. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Result carries bot commit and dirty flag, d2d binary hash, host role, load, pace
- [ ] #2 A run cut by the wall limit (no virtual-run-length-reached line) is BLOCKED, never judged
- [ ] #3 must-fail test for the cut run
<!-- AC:END -->
