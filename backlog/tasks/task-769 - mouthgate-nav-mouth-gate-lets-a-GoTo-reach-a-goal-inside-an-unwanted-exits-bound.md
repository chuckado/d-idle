---
id: TASK-769
title: 'mouthgate: nav mouth gate lets a GoTo reach a goal inside an unwanted exit''s bound'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T06:07:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 358000
card: "mouthgate"
owns:
  - "internal/nav"
exclusive: []
contracts: []
result: "act1.forgotten.countess Paladin 11 GREEN twice on the K1 fork (live, builder long lane, jobs 20261121-012324-37e7 and 012853-5d61, determinism pair same verdict, setup CHEATED = mechanics only). Fix: a GoTo waypoint inside the 15 subtile bound of an unwanted exit goes through when the walk goal lies beyond and it stays outside the 10 subtile core (nav mouth gate). Build-specific: none (nav has no class data). U: whether a click 10 to 15 from stairs fires in the Tower Cellars."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 must-fail tests Paladin and Sorceress start, real Tower Cellar 5 geometry
- [x] #2 scenario act1.forgotten.countess Paladin 11 past the gate (live)
<!-- AC:END -->
