---
id: TASK-747
title: 'mgcache: cached whole-repo gate measured over 20 merges then default; nightly uncached race with red hold'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:49:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 352000
card: "mgcache"
owns:
  - "docs/process/pg-ops-mgcache.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design 4.2 and 4.3, 7 card 12. Private ops repo work. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Measure cached vs -count=1 gate over 20 merges
- [ ] #2 drop -count=1 from the merge gate afterwards (Q3)
- [ ] #3 nightly uncached whole-repo -race with red hold
<!-- AC:END -->
