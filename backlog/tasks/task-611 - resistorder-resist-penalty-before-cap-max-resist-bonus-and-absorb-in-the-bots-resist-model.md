---
id: TASK-611
title: 'resistorder: resist penalty before cap, max-resist bonus and absorb in the bot''s resist model'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T11:06:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 308000
card: "resistorder"
owns:
  - "docs/research/resistorder.md"
  - "internal/d2/resist.go"
  - "internal/combat/rules.go"
  - "internal/readiness/model.go"
  - "internal/power/margin.go"
exclusive: []
contracts: []
result: "Resist model now adds the penalty before the cap, cap 75 + max bonus, percent absorb folded when the stat is present; offline tests only, not verified live (stat exposure U)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

decomp37 B1 and T1. Done when the bot applies game.exe's order.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Every site that computes effective resist listed with file:line
- [x] #2 Penalty added before the cap (75 + max bonus), floor -100, table in data (d2.ResistTable)
- [x] #3 Percent absorb folded into intake as a data flag, default no change without the stat
- [x] #4 Tests on two classes and three difficulties; readiness goldens updated and explained
- [x] #5 Docs updated
<!-- AC:END -->
