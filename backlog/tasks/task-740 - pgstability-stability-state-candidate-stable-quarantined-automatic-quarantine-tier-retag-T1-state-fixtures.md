---
id: TASK-740
title: 'pgstability: stability state candidate/stable/quarantined, automatic quarantine, tier retag, T1 state fixtures'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:47:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 346000
card: "pgstability"
owns:
  - "internal/scenario/stability.go"
  - "tools/provingground/scenarios/tiers.json"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design 2.7 and 3, 7 card 6 (Q7: auto-quarantine, a person releases). Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 5 equal runs make stable; a flip on unchanged commit and d2d hash quarantines by tool
- [ ] #2 tier per scenario from the catalogue
- [ ] #3 must-fail test for the flip rule
<!-- AC:END -->
