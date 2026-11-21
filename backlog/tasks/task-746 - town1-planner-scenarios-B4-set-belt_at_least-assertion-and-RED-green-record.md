---
id: TASK-746
title: 'town1: planner scenarios (B4 set, belt_at_least assertion) and RED/green record'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-21T04:49:00Z'
labels:
  - bot
  - drill
dependencies:
  - TASK-715
priority: medium
ordinal: 341000
card: "town1"
owns:
  - "tools/provingground/scenarios/townplan.json"
  - "internal/scenario/state.go"
  - "internal/scenario/newkinds_test.go"
  - "docs/probes/town-plan-scenarios.md"
exclusive: []
contracts: []
result: "The 12 B4 scenarios are in tools/provingground/scenarios/townplan.json (Paladin and Sorceress matrix, seeds 11 and 3, no_writes) and the belt_at_least assertion exists with a unit test (go test ./internal/scenario ./cmd/d2scenario ok). NOT run: no RED/green was measured, the table in docs/probes/town-plan-scenarios.md is the design expectation; setups are CHEATED mechanics only. The scenarios set knob town.plan, which arrives with card town3."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 1 (appendix B5). Done when the scenarios load and validate offline. Source: docs/design/fable-review-2026-11-21/town-planner-b-model.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 scenario files for the 12 B4 sketches, each on the Paladin and Sorceress matrix with no_writes
- [x] #2 belt_at_least {kind,min} assertion with a unit test
- [x] #3 Expected RED/green per scenario recorded in docs/probes/town-plan-scenarios.md; NOT run on the builder
<!-- AC:END -->
