---
id: TASK-560
title: 'survfix: A/B screen analysis, abreport horizon, one clock in survival wiring'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-16T19:05:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 293000
card: "survfix"
owns:
  - "docs/survival/ab-screen-analysis-2026-11-16.md"
  - "tools/survival/abreport"
  - "cmd/d2bot/survivalwire.go"
  - "internal/survival/survival_test.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/live.go"
  - "cmd/d2bot/powerwire_test.go"
exclusive: []
contracts: []
result: "Offline, not verified live. abreport now reads the last minute line (2 h runs have 119 intervals); the survival wiring used time.Now() against lockstep epoch event stamps, so the ledger never saw a chicken, death or merc death (k 1.00 in all 91 logged choices): fixed with one clock. Analysis and proposals in docs/survival/ab-screen-analysis-2026-11-16.md; bot.zoneModel stays 0; Paladin rescreen worthwhile, Sorceress after yield and mana fixes."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Fix the A/B harness horizon and the clock mix that blinded the ledger in lockstep; analyse the screen; propose model fixes. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 abreport reads the last minute line at the horizon, tested
- [x] #2 ledger and controller use one clock (lockstep epoch clock)
- [x] #3 analysis doc with numbers and proposals
- [x] #4 rescreen readiness stated
<!-- AC:END -->
