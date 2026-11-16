---
id: TASK-534
title: 'dmgcheck: why a surrounded hero took no damage (Den of Evil)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-16T03:05:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 284000
card: "dmgcheck"
owns:
  - "docs/research/damage-intake-2026-11-15.md"
  - "tools/emu/dmg_model.py"
exclusive: []
contracts: []
result: "no regression found; sandbox build lacks upstream PRs #515-#557; offline d2d scenario NOT run (builder unreachable, data only on the worker host); not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: user report of a zero-damage fight in a crowd. Done when: the trace doc has the live numbers, build ancestry and verdict. Consumer: fidelity-matrix and the intake card.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Trace doc with live Den numbers, build ancestry, rule expectation, verdict
- [x] #2 Drill recipe for the offline scenario on the builder
<!-- AC:END -->
