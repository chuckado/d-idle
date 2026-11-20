---
id: TASK-687
title: 'survhold7: realized-yield starve and unvisited-prior discount for the zone model'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T14:37:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 335000
card: "survhold7"
owns:
  - "internal/survival/"
  - "docs/probes/survival-holdout4-2026-11-19.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 yieldStarved and the bounded unvisited discount are data in params.json, off except the caster row; unchanged with knobs off
- [ ] #2 must-fail tests on Sorceress and Paladin builds, race count 20 clean
- [ ] #3 Sorceress screen on the builder long lane, addendum written
<!-- AC:END -->
