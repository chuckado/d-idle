---
id: TASK-946
title: 'epic-class-druid: Druid playable through Normal'
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
ordinal: 447000
card: "epic-class-druid"
owns:
  - "docs/epics/epic-class-druid.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Class epic for the Druid. Builds in internal/plandata: build-druid_fire_normal30.json, build-druid_fire_guide.json, build-druid_tornado_endgame90.json; unlock table unlocks-druid.json. Docs: docs/classes/druid.md. What exists today: no drill rows, no live run; ground-target casts are the open design item (ADR 0020 class order). Blockers: None known beyond: no drill rows. Living status: docs/epics/epic-class-druid.md. Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Druid: class traits, plan table (internal/plandata) and combat profile (internal/combat/classdata) complete and data-only, no class literals in Go; a second-build test passes (build flexibility rule).
- [ ] #2 Druid: leveling plan clvl 1 to 30+ per the XP table with survival gates (plandata build-*normal30.json), drilled to clvl 30 in the lab.
- [ ] #3 Druid: stat and skill allocation plan locked and tested (autoalloc, capability table), spends every point.
- [ ] #4 Druid: kit and economy profile (belt, potions, merc) for the class, no death spiral in a 3 h run.
- [ ] #5 Druid: quest drills GREEN as Druid for every act (drill matrix rows with -class Druid; today: no drill rows, no live run; ground-target casts are the open design item (ADR 0020 class order)).
- [ ] #6 Druid: a fresh-hero live Normal run (no cheats, writes after start 0) reaches the end of Act N; record N and the real clvl (today: no uncheated live run recorded here, U).
- [ ] #7 Druid: known blockers listed and closed. Current: None known beyond: no drill rows.
<!-- AC:END -->
