---
id: TASK-746
title: 'town1: planner scenarios (B4 set, belt_at_least assertion) and RED/green record'
status: Proposed
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
  - "tools/provingground/scenarios/town.plan.*"
  - "internal/scenario/state.go"
  - "docs/probes/town-plan-scenarios.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 1 (appendix B5). Done when the scenarios load and validate offline. Source: docs/design/fable-review-2026-11-21/town-planner-b-model.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 scenario files for the 12 B4 sketches, each on the Paladin and Sorceress matrix with no_writes
- [ ] #2 belt_at_least {kind,min} assertion with a unit test
- [ ] #3 RED/green on today's main recorded in a probe doc (not run if the builder is busy; say so)
<!-- AC:END -->
