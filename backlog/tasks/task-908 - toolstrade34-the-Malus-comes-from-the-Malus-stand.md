---
id: TASK-908
title: 'toolstrade34: the Malus comes from the Malus stand'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T23:19:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 413000
card: "toolstrade34"
owns:
  - "internal/task/act1.go"
  - "internal/bot/campsim_objects_test.go"
  - "internal/task/steps.go"
  - "internal/task/sim/sim.go"
  - "internal/task/gear_gate_test.go"
  - "internal/quest/dialog.go"
  - "cmd/d2bot/objects.go"
  - "cmd/d2bot/objects_layout_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "The malus step is StepStandPickUp (click the stand object 108, then pick up); the sim gives hdm only from the stand; unit-tested; the lab hand-in run is not done; whether the real game drops the Malus on the Smith death is unverified."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Fork agent finding (V d2d code and lab): hdm drops only from a click on the Malus stand. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The Malus step clicks the Malus stand (object 108, click until its mode changes) and then picks the Malus up; the fake game gives the Malus only from the stand; must-fail test, paladin and necromancer
<!-- AC:END -->
