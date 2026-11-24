---
id: TASK-917
title: 'flip: Necromancer step scenarios rerun after toolstrade40 (#1249)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T04:07:00Z'
labels:
  - bot
  - drill
dependencies: []
priority: medium
ordinal: 421000
card: "flip"
owns:
  - "data/stepgrounds/verdicts.json"
  - "docs/probes/quest-step-catalogue.md"
exclusive: []
contracts: []
result: "Rerun on b74549d93 (devp26, job 20261123-225912-cdf0): nothing flipped RED to GREEN. Loop guard works (parks after 12 repeats). room_stand and pickup stay RED (plan walks the hero out before a chunk exists); stand_click_afar went GREEN to RED (stand step skipped after 8 clicks; flake or regression, U); caster_loop and plan_overlap were never RED. CHEATED, mechanics only."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Five Necromancer step scenarios rerun on main with #1249, flips recorded
<!-- AC:END -->
