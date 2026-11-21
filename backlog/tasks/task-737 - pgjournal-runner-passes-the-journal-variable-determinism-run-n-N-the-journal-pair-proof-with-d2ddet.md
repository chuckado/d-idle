---
id: TASK-737
title: 'pgjournal: runner passes the journal variable; determinism-run -n N; the journal pair proof with d2ddet'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:46:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 343000
card: "pgjournal"
owns:
  - "cmd/d2scenario/journal.go"
  - "docs/probes/pg-journal-pair-2026-11-21.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design section 2.5, 7 card 3. About 40 builder minutes. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Runner passes D2DCTL_JOURNAL for the bot
- [ ] #2 determinism-run -n N writes N runs and compares them
- [ ] #3 Pair of design 2.5 (town.resupply and act1.burial.blood_raven, Paladin 11, 3 runs each) run on the builder main lane when free and written up with d2ddet compare output
<!-- AC:END -->
