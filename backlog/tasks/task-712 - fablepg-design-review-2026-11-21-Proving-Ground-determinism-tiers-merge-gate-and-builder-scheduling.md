---
id: TASK-712
title: 'fablepg: design review 2026-11-21, Proving Ground determinism, tiers, merge gate and builder scheduling'
status: Proposed
assignee:
  - user
created_date: '2026-11-21T03:15:00Z'
labels:
  - docs
  - research
dependencies: []
priority: high
ordinal: 340000
card: "fablepg"
owns:
  - "docs/design/fable-review-2026-11-21/proving-ground-gate.md"
  - "docs/design/fable-review-2026-11-21/proving-ground-gate-catalog.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Owner-requested Fable design review of the Proving Ground, the merge gate and builder scheduling. Docs only, nothing changes Go, the merge scripts or the builder. Consumer: the owner (numbered questions in the doc) and the rollout cards it proposes. Done when the docs are merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 proving-ground-gate.md written: determinism root causes, tiers, merge gate, results store, harness hygiene, rollout cards, questions
- [ ] #2 proving-ground-gate-catalog.md written: scenario catalogue and coverage map with setup honesty tags
- [ ] #3 contracts, backlog and leakcheck checks pass
<!-- AC:END -->
