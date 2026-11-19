---
id: TASK-673
title: 'agencyp0: Agency P0 stat and skill points by verb, sheet, autopilot switch, legality'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T18:45:00Z'
labels:
  - bot
  - feature
  - T0
dependencies: []
priority: medium
ordinal: 329000
card: "agencyp0"
owns:
  - "internal/points"
  - "internal/watch/pointsapi.go"
  - "internal/watch/web/points.js"
  - "cmd/d2bot/pointswire.go"
  - "docs/points.md"
  - "tools/provingground/scenarios/points.json"
  - "docs/process/points-verbs-devp14.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Roadmap R3 of the fable review. Done when the checks pass and the branch is merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 spend stat and skill points by d2d verb behind a capability check, read back
- [ ] #2 points sheet API and Points tile; POSTs audited
- [ ] #3 per-hero autopilot switch default off, points waiting state
- [ ] #4 legality fixes: skill cap check, ladder runeword tags, missing skill kinds, resist order
- [ ] #5 Proving Ground scenarios for points
<!-- AC:END -->
