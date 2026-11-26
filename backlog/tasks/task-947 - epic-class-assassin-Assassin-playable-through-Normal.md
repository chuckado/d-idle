---
id: TASK-947
title: 'epic-class-assassin: Assassin playable through Normal'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:05:00Z'
labels:
  - bot
  - feature
  - T4
dependencies: []
priority: medium
ordinal: 448000
card: "epic-class-assassin"
owns:
  - "docs/epics/epic-class-assassin.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Class epic for the Assassin. Builds in internal/plandata: build-assassin_fireblast_normal30.json, build-assassin_sentry_endgame90.json; unlock table unlocks-assassin.json. Docs: docs/classes/assassin.md. What exists today: no drill rows, no live run; needs a trap planner (docs/classes/assassin.md). Blockers: None known beyond: no drill rows. Living status: docs/epics/epic-class-assassin.md. Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Assassin: class traits, plan table (internal/plandata) and combat profile (internal/combat/classdata) complete and data-only, no class literals in Go; a second-build test passes (build flexibility rule).
- [ ] #2 Assassin: leveling plan clvl 1 to 30+ per the XP table with survival gates (plandata build-*normal30.json), drilled to clvl 30 in the lab.
- [ ] #3 Assassin: stat and skill allocation plan locked and tested (autoalloc, capability table), spends every point.
- [ ] #4 Assassin: kit and economy profile (belt, potions, merc) for the class, no death spiral in a 3 h run.
- [ ] #5 Assassin: quest drills GREEN as Assassin for every act (drill matrix rows with -class Assassin; today: no drill rows, no live run; needs a trap planner (docs/classes/assassin.md)).
- [ ] #6 Assassin: a fresh-hero live Normal run (no cheats, writes after start 0) reaches the end of Act N; record N and the real clvl (today: no uncheated live run recorded here, U).
- [ ] #7 Assassin: known blockers listed and closed. Current: None known beyond: no drill rows.
<!-- AC:END -->
