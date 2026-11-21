---
id: TASK-742
title: 'pgselfcheck: hero self-check on the prep d2d and its static half in validate'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:48:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 348000
card: "pgselfcheck"
owns:
  - "internal/scenario/hero/selfcheck.go"
  - "cmd/d2scenario/validate.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design section 6.2, 7 card 8. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Self-check refuses a hero whose setup tags or state do not match the profile
- [ ] #2 static half runs in validate offline
- [ ] #3 must-fail test
<!-- AC:END -->
