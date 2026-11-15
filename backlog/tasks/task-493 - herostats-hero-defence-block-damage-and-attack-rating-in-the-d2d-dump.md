---
id: TASK-493
title: 'herostats: hero defence, block, damage and attack rating in the d2d dump'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T11:53:00Z'
updated_date: '2026-11-15T12:21:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 274000
card: "herostats"
owns:
  - "docs/survival/herostats-2026-11-15.md"
  - "internal/d2dsrc/gearstats.go"
  - "internal/d2dsrc/gearstats_test.go"
  - "internal/power/hero.go"
  - "docs/survival/power.md"
exclusive: []
contracts: []
result: "Fork branch sandbox-herostats (tip 02ceea54, pushed) adds stats.gear, stats.weapons and stats.hero (block, ar, damage per hand, dmgpct); Go parser overlays hero additively. test_hero_stats and test_gear_stats pass on the builder; full d2d link and live panel check not done (not verified live)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fork dump stats carries gear, weapons and hero (block, ar, damage per hand, dmgpct), pushed to the private fork
- [x] #2 Go parser reads hero additively; tests pass
- [x] #3 Docs: power.md readable table, research note with V/U tags
<!-- AC:END -->
