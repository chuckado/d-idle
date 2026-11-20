---
id: TASK-686
title: 'survhold6: Sorceress hero gets a wand, rerun 211 213 241 243'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T14:17:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 334000
card: "survhold6"
owns:
  - "tools/survival/abrun.sh"
  - "docs/probes/survival-holdout4-2026-11-19.md"
result: "Sorceress A/B hero now wears a wand (wnd, SETUP, both arms); Paladin always had hax. Rerun 211 213 (builder) 241 243 (second worker), writes after start 0: pooled margin 29123 vs legacy 73274 xp (ratio 0.397), chickens 1/0, deaths 0/0: FAIL on yield (abreport INCONCLUSIVE at 2 pairs per host). Cause is the model: zone yield 150 xp/min modelled vs 4-27 realized, same zone re-picked, not a weapon defect. No default changed; Paladin unchanged"
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
- [x] #1 Sorceress A/B hero gets a wand as SETUP in both arms; Paladin weapon verified
- [x] #2 Sorceress seeds 211 213 (builder) and 241 243 (second worker) rerun, writes after start 0
- [x] #3 addendum with corrected verdict in the holdout 4 doc
<!-- AC:END -->
