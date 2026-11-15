---
id: TASK-508
title: 'powerwire: wire decomp29 skill damage into the power model (shadow)'
status: Proposed
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
- [ ] #1 skilldamage.json loader and OffenceDPS estimator (Paladin Zeal and Sorceress Fire Bolt tested, missing data flagged)
- [ ] #2 swingSeconds, defaultDpsPerClvl, manaPerSecond derived from the data when the build is known; defaults kept flagged
- [ ] #3 replay before/after reported (k_hero, planner agreement, pass rule, risky minutes)
- [ ] #4 docs/survival/power.md readable-today table updated
<!-- AC:END -->
