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
result: "Design review docs/design/fable-review-2026-11-21/proving-ground-gate.md and -catalog.md: the digest differences are traced to harness and bot wall-clock inputs (wall-clock board tapper, whole-state digest, goal and farm timers), by code reading, NOT proven by a run; assertion digest and 5-run stability rule with automatic quarantine; tiers T0 to T4 and which event each gates; merge gate measured from the queue log (gates median 435 s, p90 1990 s) with a cached one-job gate, merge priority, agent quotas and a worker watchdog; one results ledger and board page; hero library with a setup self-check; 16 rollout cards; 10 questions for the owner. Docs only, nothing run live, no Go or script changed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Owner-requested Fable design review of the Proving Ground, the merge gate and builder scheduling. Docs only, nothing changes Go, the merge scripts or the builder. Consumer: the owner (numbered questions in the doc) and the rollout cards it proposes. Done when the docs are merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 proving-ground-gate.md written: determinism root causes, tiers, merge gate, results store, harness hygiene, rollout cards, questions
- [x] #2 proving-ground-gate-catalog.md written: scenario catalogue and coverage map with setup honesty tags
- [x] #3 contracts, backlog and leakcheck checks pass
<!-- AC:END -->
