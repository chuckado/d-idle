---
id: TASK-916
title: >-
  redsteps: Necromancer step scenarios built to go RED (room start, caster belt
  loop, plan overlap) and pre.warp
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T03:46:00Z'
updated_date: '2026-11-26T01:45:00Z'
labels:
  - bot
  - drill
dependencies: []
priority: medium
ordinal: 420000
card: "redsteps"
owns:
  - "tools/provingground/scenarios/stepgrounds.json"
  - "data/stepgrounds/steps.json"
  - "data/stepgrounds/verdicts.json"
  - "docs/probes/quest-step-catalogue.md"
  - "internal/scenario/scenario.go"
  - "internal/scenario/setup.go"
  - "internal/scenario/necro_test.go"
exclusive: []
contracts: []
result: "pre.warp (debug warp x y, tile cells) with a unit test; three scenarios on devp26: room_stand RED (plan walks the hero out of the Barracks 24 s before the chunk starts), caster_loop and plan_overlap GREEN (conditions not reproduced). Rerun after toolstrade40 pending. CHEATED, mechanics only."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 pre.warp with a unit test
- [x] #2 Three scenarios and verdicts on devp26
<!-- AC:END -->
