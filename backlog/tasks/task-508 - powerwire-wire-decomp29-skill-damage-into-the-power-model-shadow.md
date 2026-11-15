---
id: TASK-508
title: 'powerwire: wire decomp29 skill damage into the power model (shadow)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T15:53:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 278000
card: "powerwire"
owns:
  - "internal/power/"
  - "docs/survival/power.md"
  - "tools/levelsim/"
  - "cmd/d2bot/powerwire.go"
  - "cmd/d2bot/survivalwire.go"
exclusive: []
result: "Skill damage data is read by internal/power (Estimate, OffenceDPS, flags) and used by Margin in shadow only; swing, caster dps and mana per second come from the data when build and class are known, constants stay flagged fallbacks. Replay on the sandbox log: no change without a weapon; with an ASSUMED weapon the pass rule gets worse (1 of 14) and k_hero 3.00 (n=1, nothing tuned). Verified offline (unit tests on Paladin Zeal and Sorceress Fire Bolt, replay on the worker host); not verified live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: decomp29 found the survival model swing 2x too slow and caster dps and mana constants far off. Done when: the data drives the estimate in shadow, flagged, with the replay before and after reported.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 skilldamage.json loader and OffenceDPS estimator (Paladin Zeal and Sorceress Fire Bolt tested, missing data flagged)
- [x] #2 swingSeconds, defaultDpsPerClvl, manaPerSecond derived from the data when the build is known; defaults kept flagged
- [x] #3 replay before/after reported (k_hero, planner agreement, pass rule, risky minutes)
- [x] #4 docs/survival/power.md readable-today table updated
<!-- AC:END -->
