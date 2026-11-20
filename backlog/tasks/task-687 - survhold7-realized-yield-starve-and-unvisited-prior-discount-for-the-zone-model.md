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
result: "yieldStarved and a bounded unvisited-prior discount as caster-row data (global off, zoneModel untouched); tests on two builds, race x20 clean; Sorceress screen 4 pairs on the builder: ratio 0.615 against 0.397 before, INCONCLUSIVE, entry gate untouched, no flip. Verified live (builder runs), writes after start 0."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the zone model keeps a zero-yield zone and chases unvisited priors (holdout4 addendum 5). Done when: caster row data plus tests plus screen, entry gate untouched.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 yieldStarved and the bounded unvisited discount are data in params.json, off except the caster row; unchanged with knobs off
- [x] #2 must-fail tests on Sorceress and Paladin builds, race count 20 clean
- [x] #3 Sorceress screen on the builder long lane, addendum written
<!-- AC:END -->
