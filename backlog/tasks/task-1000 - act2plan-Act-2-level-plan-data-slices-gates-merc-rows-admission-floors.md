---
id: TASK-1000
title: 'act2plan: Act 2 level plan data (slices, gates, merc rows, admission floors)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T17:41:00Z'
labels:
  - docs
  - feature
  - T4
dependencies: []
priority: medium
ordinal: 482000
card: "act2plan"
lane: next
epic: epic-act2
owns:
  - "internal/plandata/act2plan.json"
  - "internal/plandata/act2plan.go"
  - "internal/plandata/act2plan_test.go"
  - "internal/plandata/admission.json"
  - "docs/plan/data/act2plan.json"
  - "docs/plan/data/admission.json"
  - "docs/plan/act2-level-plan.md"
  - "docs/epics/epic-act2.md"
exclusive: []
contracts: []
result: "Act 2 plan data for two classes, eight dungeon road rows and the SME answer (S) are in; tests pass on the worker host; not verified live; XP rows and Act 2 spine chapters still missing"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Why: epic-act2 blocker no Act 2 level plan. Done when data, doc and tests are in.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Act 2 plan data for paladin and sorceress validated by table-driven test
- [x] #2 admission road rows for Act 2 dungeons, copies equal (TestMatchesDocs)
- [x] #3 SME answer recorded verbatim tagged S in docs/plan/act2-level-plan.md
<!-- AC:END -->
