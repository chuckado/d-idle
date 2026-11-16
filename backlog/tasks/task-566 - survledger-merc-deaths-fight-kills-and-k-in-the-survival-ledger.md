---
id: TASK-566
title: 'survledger: merc deaths, fight kills and k in the survival ledger'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T21:08:00Z'
updated_date: '2026-11-16T21:53:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 295000
card: "survledger"
owns:
  - "internal/survival"
  - "cmd/d2bot/survivalwire.go"
  - "cmd/d2bot/senseshistory.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/live.go"
exclusive: []
contracts: []
result: "Offline only, not verified live. Causes: PowerOf counted the merc (x1.25), so his death moved power 20 percent and the next decision wiped the learned row; the fight hook rode useSensesHistory, which returns early without a senses rig (d2d harness), so kills/seconds stayed 0. Fixed both, fights without a prediction add kills/seconds only, k printed on every zone choice line. Build-agnostic; tested on Paladin and Sorceress."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the rescreen found merc deaths and fight kills/seconds at 0 in the ledger, so k could not learn. Done when: both reach the ledger, k prints on every zone choice line, two-class test passes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Merc deaths survive the next decision (power scalar ignores the merc)
- [x] #2 Fight kills and seconds reach the ledger without a senses rig
- [x] #3 k printed on every zone choice line
- [x] #4 Two-class test with epoch clock
<!-- AC:END -->
