---
id: TASK-749
title: 'mglanes: merge class, agent quotas, whole-repo refusal, wait budget, load guard'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
labels:
  - ops
  - feature
dependencies: []
priority: medium
ordinal: 353000
card: "mglanes"
owns:
  - "docs/process/pg-ops-mglanes.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design 4.5, 7 card 13. Private ops repo work. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Quotas: 1 running + 1 waiting main-lane job per agent, 2 on the long lane, no whole-repo from a branch (Q5)
- [ ] #2 long lane two slots behind the memory guard (Q9)
- [ ] #3 load guard and wait budget
<!-- AC:END -->
