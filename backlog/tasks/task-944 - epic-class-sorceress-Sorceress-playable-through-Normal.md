---
id: TASK-944
title: 'epic-class-sorceress: Sorceress playable through Normal'
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
ordinal: 445000
card: "epic-class-sorceress"
owns:
  - "docs/epics/epic-class-sorceress.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Class epic for the Sorceress. Builds in internal/plandata: build-sorceress_normal30.json, build-sorceress_fire_normal30.json, build-sorceress_blizzard_endgame90.json, build-fallback-lightning-v3.json, build-primary-blizzard.json; unlock table unlocks-sorceress.json. Docs: docs/sorcnormal30.md, docs/wizardclass.md. What exists today: the default class (fallback lightning, normal30 table); Sail east drill GREEN as Sorceress (matrix note 2); the other quest rows are drilled as Paladin. Blockers: Rows other than sail east are not drilled as Sorceress; Act 1 Tools of the Trade RED applies. Living status: docs/epics/epic-class-sorceress.md. Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Sorceress: class traits, plan table (internal/plandata) and combat profile (internal/combat/classdata) complete and data-only, no class literals in Go; a second-build test passes (build flexibility rule).
- [ ] #2 Sorceress: leveling plan clvl 1 to 30+ per the XP table with survival gates (plandata build-*normal30.json), drilled to clvl 30 in the lab.
- [ ] #3 Sorceress: stat and skill allocation plan locked and tested (autoalloc, capability table), spends every point.
- [ ] #4 Sorceress: kit and economy profile (belt, potions, merc) for the class, no death spiral in a 3 h run.
- [ ] #5 Sorceress: quest drills GREEN as Sorceress for every act (drill matrix rows with -class Sorceress; today: the default class (fallback lightning, normal30 table); Sail east drill GREEN as Sorceress (matrix note 2); the other quest rows are drilled as Paladin).
- [ ] #6 Sorceress: a fresh-hero live Normal run (no cheats, writes after start 0) reaches the end of Act N; record N and the real clvl (today: no uncheated live run recorded here, U).
- [ ] #7 Sorceress: known blockers listed and closed. Current: Rows other than sail east are not drilled as Sorceress; Act 1 Tools of the Trade RED applies.
<!-- AC:END -->
