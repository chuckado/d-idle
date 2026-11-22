---
id: TASK-799
title: 'fogtrail: Map panel shows walked ground and the objective marker, forgets it per game'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T02:35:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 367000
card: "fogtrail"
owns:
  - "internal/senses/objective.go"
  - "internal/senses/objective_test.go"
  - "internal/watch/web/panels/map"
  - "internal/watch/testdata/map_panel.js"
  - "internal/watch/senses.go"
  - "internal/watch/senses_test.go"
  - "internal/nav/roam.go"
  - "internal/nav/objectiveview_test.go"
  - "cmd/d2bot/sensesfeed.go"
  - "docs/mapfog.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: review docs/design/fable-review-2026-11-22/map-knowledge.md card 5 and Q5. Done when the viewer shows walked ground plus the objective marker.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 objective key in /api/senses (not fogged), current area only
- [ ] #2 Map panel draws the objective ring over the fog, phone-friendly, no inner scroll box
- [ ] #3 walked ground forgotten per game (Publisher.NewGame via ForgetCoverage); size and persistence decided in docs/mapfog.md
- [ ] #4 must-fail UI test (map_panel.js) and Go tests
<!-- AC:END -->
