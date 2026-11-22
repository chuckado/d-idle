---
id: TASK-798
title: 'knownedge: FullyKnown is measured; a whole grid is planned on alone (knob nav.fullGrid)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T02:31:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 367000
card: "knownedge"
owns:
  - "internal/nav/fullgrid_test.go"
  - "internal/nav/roam.go"
  - "internal/nav/nav.go"
  - "internal/env/tuning_nav.go"
  - "internal/d2dworld/world.go"
  - "docs/knobs.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: review docs/design/fable-review-2026-11-22/map-knowledge.md section 5 card 2, after the gridwhole probe. Done when the tests pass and the knob is documented.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 FullyKnown is measured from the grid (known cells equal the level rectangle) behind nav.fullGrid (default 0)
- [ ] #2 A whole grid refuses a goal outside it instead of walking to a known edge; a partial grid keeps the optimistic walk
- [ ] #3 Must-fail tests for the Paladin and Sorceress walks (internal/nav/fullgrid_test.go)
- [ ] #4 d2dworld logs the known cell count of every arrival
<!-- AC:END -->
