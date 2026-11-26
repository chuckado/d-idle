---
id: TASK-988
title: 'inputl4: Input layer I4: fork verbs buy, sell, repair, take-waypoint, menu-choos'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:58:00Z'
labels:
  - d2d
  - feature
dependencies:
  - TASK-986
priority: medium
ordinal: 475000
card: "inputl4"
owns:
  - "docs/design/input-i4.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Recommendation R17 of the 2026-11-19 review (07 table I4; ADR 0021). Fork branch work: one verb per game client message with a differential each. Needs `exclusive: fork-branch:<name>` and the same-day push rule in AGENTS.md; the implementing agent sets them when it starts. Patches never go into this repo. Risk: medium. Enable criteria: one differential per verb.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fork verbs with a differential each, pushed to the private remote the same day
<!-- AC:END -->
