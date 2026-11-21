---
id: TASK-753
title: 'town3: supply.NeedFor, Solve, profiles.json, town.plan stage 1 shadow trace'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:50:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-715
priority: medium
ordinal: 343000
card: "town3"
owns:
  - "internal/supply"
  - "data/supply"
  - "internal/env/tuning_town.go"
  - "docs/knobs.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Town planner card 3 (B5), stage 1.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 data/supply/profiles.json (today's constants), NeedFor, pure Solve with the 4.2 priority table; errors name profile/build/line
- [ ] #2 knob town.plan (0 off, 1 shadow ...) default 0; stage 1 logs 'town plan (shadow): needs, has, would buy' and the minute-line field; no behaviour change
- [ ] #3 Solve tests on Paladin and Sorceress; determinism test (T8)
<!-- AC:END -->
