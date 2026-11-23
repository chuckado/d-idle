---
id: TASK-892
title: 'toolstrade20: the fight walks toward a target behind doors before it ignores it'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T11:40:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 401000
card: "toolstrade20"
owns:
  - "internal/combat/fight.go"
  - "internal/combat/sim_test.go"
  - "internal/combat/reach_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "The fight calls GoTo (door-augmented since #1188) toward an unreachable target up to 3 times before ignoring it; unit-tested; lab and live effect not seen."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Necromancer lab: the Smith's room was behind unopened doors and the fight wrote him off. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A target with no path and no firing position gets up to three door-augmented walks (GoTo) before it is ignored as unreachable; must-fail test
<!-- AC:END -->
