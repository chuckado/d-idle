---
id: TASK-684
title: 'survhold4: holdout 4, Sorceress screen and Paladin seeds on valid heroes'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-20T04:52:00Z'
updated_date: '2026-11-20T06:41:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 333000
card: "survhold4"
owns:
  - "tools/survival"
  - "docs/probes/survival-holdout4-2026-11-19.md"
result: "Harness fix: abrun sets resource.pointsDecisions 1 (agencyp0 points gate left heroes unspent; first wave invalid). Sorceress 4 pairs (107 181 211 213, builder, writes after start 0): XP ratio 0.669, chickens 4/1, 0 deaths, INCONCLUSIVE and 211/213 margin low-yield, wave 4 not run. Paladin 2 fresh seeds (221 223) ratio 1.085, no deaths or chickens. No default changed; recommendation in the doc"
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
- [x] #1 Sorceress 3-pair screen and 2 more Paladin seeds on the builder long lane, writes after start 0
- [x] #2 abrun sets resource.pointsDecisions 1 (heroes spend points after agencyp0)
- [x] #3 docs/probes/survival-holdout4-2026-11-19.md with per-class verdict
<!-- AC:END -->
