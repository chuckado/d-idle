---
id: TASK-714
title: 'fableboard: Fable design review of the quest board state machine'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T03:17:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 340000
card: "fableboard"
owns:
  - "docs/design/fable-review-2026-11-21/quest-board*"
exclusive: []
contracts: []
result: "Docs only, nothing verified live and no test run: quest-board.md plus appendices A (file:line evidence, read at 5b12ce5e7 and e77b53911) and B (reason codes, API, invariants and scenarios, panel spec). Root cause: the board stores no lifecycle and derives state from non-monotonic evidence. Proposal: stored runs, completion latch, reason codes; 10 migration cards; 11 questions for the user. No build-specific part."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: six related quest board bugs in one day (two accepted chunks parked, Stop on a queued chunk, prerequisite deadlock, guard hold without restock, evidence regression, no visible reason). Done when: docs/design/fable-review-2026-11-21/quest-board.md and its appendices give the shared cause, a decision-ready model, the migration cards and the questions for the user. Docs only. Consumer: boardprereq (the in-flight prerequisite fix) and the migration cards listed in section 5 of the doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Root-cause analysis with file:line evidence for the six board bugs of 2026-11-20
- [x] #2 Clean model: chunk and quest state machine, dependency semantics, invariants, reason codes, API semantics, regression, persistence
- [x] #3 Arbiter relation, migration cards, panel spec, numbered open questions with a recommendation
<!-- AC:END -->
