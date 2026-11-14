---
id: TASK-460
title: 'survcal: survival calibration from the sandbox replay (S-B2/S-C2)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T23:55:00Z'
labels:
  - bot
  - model-check
dependencies: []
priority: medium
ordinal: 263000
card: "survcal"
owns:
  - "docs/survival/calibration-2026-11-14.md"
  - "internal/survival"
  - "internal/power/hero.go"
  - "internal/bot/lifelog.go"
  - "internal/bot/lifelog_test.go"
  - "docs/lifelog.md"
  - "tools/levelsim/shadow.go"
  - "tools/levelsim/controller.go"
  - "tools/levelsim/shadow_test.go"
  - "cmd/d2bot/survivalwire.go"
exclusive: []
contracts: []
result: "Offline, not live. Constants as data (predTakenFloorPerKill 1, kMax 3, mercPriorMinutes 25, mercWindowMinutes 120 for the either row, kHeroUnreadMax 2) with tests on Paladin and Sorceress; zone entry now logs def, block, dmg, dmgpct, ar (logging only). Replay on the Tower builder, balanced: k_hero 6.2 to 2.0, Cave Level 1 chosen 163 to 17 of 445 minutes, risky 249 to 86, planner agreement 2 to 5 of 45, structural checks unchanged. S-B pass rule still FAILS 6 of 14 (unchanged until logs carry the new fields); n=1, not tuned to pass."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the replay (card survreplay) showed the S-C controller far too timid on a log without weapon or defence terms. Done when: the proposed fixes are data with tests, the zone entry line carries the missing terms, and the replay is rerun and reported (docs/survival/calibration-2026-11-14.md). Consumer: S-D.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Calibration constants are data in params.json with tags and tests
- [x] #2 zone entry line carries def, block, dmg, dmgpct, ar (logging only), tested on two classes
- [x] #3 Replay rerun before/after reported
<!-- AC:END -->
