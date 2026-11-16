---
id: TASK-554
title: 'survabscreen: survival A/B 3-seed screen result (bot.zoneModel 0 vs 1)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T17:38:00Z'
updated_date: '2026-11-16T19:00:00Z'
labels:
  - bot
  - d2d-measure
dependencies: []
priority: medium
ordinal: 292000
card: "survabscreen"
owns:
  - "docs/probes/survival-ab-screen-2026-11-16.md"
  - "docs/probes/test-characters.md"
exclusive: []
contracts: []
result: "Screen FAILS the rule on the Paladin (chickens 4 vs 1, slack 1; deaths 1 vs 0), Sorceress INCONCLUSIVE (XP lean lower, 0.76); 12 lockstep runs, writes after start 0 in all; not verified on game.exe. Default stays 0; 8 seeds not recommended until seed 101 and the Sorceress XP loss are read. abreport xp120 falls back to xp90 (defect, see doc)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Paladin 3-seed screen run, both arms, writes after start 0
- [x] #2 Report written with verdict and 8-seed recommendation
<!-- AC:END -->
