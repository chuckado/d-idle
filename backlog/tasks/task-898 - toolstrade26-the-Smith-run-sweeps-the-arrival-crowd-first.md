---
id: TASK-898
title: 'toolstrade26: the Smith run sweeps the arrival crowd first'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T13:18:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 406000
card: "toolstrade26"
owns:
  - "internal/boss/act1.go"
  - "internal/boss/boss_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "The Smith run now sweeps the arrival crowd before it walks to him (OuterClear, as Blood Raven); unit-tested, live effect not seen; the cause analysis is in the probe doc."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Live: the hero walked through 35 Barracks hostiles to the Smith and arrived at a quarter of its life. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Smith script has an OuterClear (radius 40, 90 s); must-fail test
<!-- AC:END -->
