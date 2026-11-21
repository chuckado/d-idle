---
id: TASK-771
title: 'survhold8: holdout 5, Sorceress 8 fresh seeds and Paladin 2'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T13:36:00Z'
updated_date: '2026-11-21T16:05:00Z'
labels:
  - bot
  - research
dependencies: []
priority: medium
ordinal: 359000
card: "survhold8"
owns:
  - "docs/probes/survival-holdout5-2026-11-21.md"
exclusive: []
contracts: []
result: "Sorceress 8 fresh pairs on the second worker: ratio 0.497, interval [-16691, -5391] under the floor, FAIL (cause: merc revive left dead and k 3.00 clamp, inferred); Paladin 2 pairs ratio 1.364, chickens 0/2, below minimum, INCONCLUSIVE. zoneModel default untouched. Verified live (the worker host), writes after start 0 in all 20 runs."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Sorceress 8 fresh pairs legacy vs margin at horizon 90, writes after start 0 each
- [x] #2 Paladin 2 fresh pairs
- [x] #3 abreport per class, per-seed table, verdict and FAIL analysis in docs/probes/survival-holdout5-2026-11-21.md
<!-- AC:END -->
