---
id: TASK-943
title: 'epic-class-barbarian: Barbarian playable through Normal'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:04:00Z'
labels:
  - bot
  - feature
  - T4
dependencies: []
priority: medium
ordinal: 444000
card: "epic-class-barbarian"
owns:
  - "docs/epics/epic-class-barbarian.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Class epic for the Barbarian. Builds in internal/plandata: build-barbarian_melee_normal30.json, build-barbarian_frenzy_guide.json, build-barbarian_warcry_endgame90.json; unlock table unlocks-barbarian.json. Docs: docs/classes/barbarian.md. What exists today: no drill rows, no live run. Blockers: None known beyond: no drill rows. Living status: docs/epics/epic-class-barbarian.md. Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Barbarian: class traits, plan table (internal/plandata) and combat profile (internal/combat/classdata) complete and data-only, no class literals in Go; a second-build test passes (build flexibility rule).
- [ ] #2 Barbarian: leveling plan clvl 1 to 30+ per the XP table with survival gates (plandata build-*normal30.json), drilled to clvl 30 in the lab.
- [ ] #3 Barbarian: stat and skill allocation plan locked and tested (autoalloc, capability table), spends every point.
- [ ] #4 Barbarian: kit and economy profile (belt, potions, merc) for the class, no death spiral in a 3 h run.
- [ ] #5 Barbarian: quest drills GREEN as Barbarian for every act (drill matrix rows with -class Barbarian; today: no drill rows, no live run).
- [ ] #6 Barbarian: a fresh-hero live Normal run (no cheats, writes after start 0) reaches the end of Act N; record N and the real clvl (today: no uncheated live run recorded here, U).
- [ ] #7 Barbarian: known blockers listed and closed. Current: None known beyond: no drill rows.
<!-- AC:END -->
