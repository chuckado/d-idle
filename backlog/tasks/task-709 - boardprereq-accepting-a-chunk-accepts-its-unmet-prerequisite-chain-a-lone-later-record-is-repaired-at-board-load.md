---
id: TASK-709
title: 'boardprereq: accepting a chunk accepts its unmet prerequisite chain, a lone later record is repaired at board load'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T02:55:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 340000
card: "boardprereq"
owns:
  - "internal/questboard"
  - "tools/provingground/scenarios/cain.json"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 accept of a chunk with unmet After steps accepts the chain in questspec order; a lone later record is repaired at board load with a log line
- [ ] #2 Stop on a prerequisite step withdraws its dependants and the toast says so
- [ ] #3 unit tests on Paladin and Sorceress; single chunk unchanged
- [ ] #4 scenarios board.prereq-not-accepted and board.prereq-start-tap run on the builder
<!-- AC:END -->
