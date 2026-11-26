---
id: TASK-941
title: 'epic-class-amazon: Amazon playable through Normal'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:03:00Z'
labels:
  - bot
  - feature
  - T4
dependencies: []
priority: medium
ordinal: 442000
card: "epic-class-amazon"
owns:
  - "docs/epics/epic-class-amazon.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Class epic for the Amazon. Builds in internal/plandata: build-amazon_bow_normal30.json, build-amazon_javelin_normal30.json, build-amazon_multishot_endgame90.json; unlock table unlocks-amazon.json. Docs: docs/classes/amazon.md, docs/classes/smoke-amazon.md. What exists today: last of the planned classes (ADR 0020 order); no quest drill rows; smoke only. Blockers: None known beyond: no drill rows, no live run. Living status: docs/epics/epic-class-amazon.md. Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Amazon: class traits, plan table (internal/plandata) and combat profile (internal/combat/classdata) complete and data-only, no class literals in Go; a second-build test passes (build flexibility rule).
- [ ] #2 Amazon: leveling plan clvl 1 to 30+ per the XP table with survival gates (plandata build-*normal30.json), drilled to clvl 30 in the lab.
- [ ] #3 Amazon: stat and skill allocation plan locked and tested (autoalloc, capability table), spends every point.
- [ ] #4 Amazon: kit and economy profile (belt, potions, merc) for the class, no death spiral in a 3 h run.
- [ ] #5 Amazon: quest drills GREEN as Amazon for every act (drill matrix rows with -class Amazon; today: last of the planned classes (ADR 0020 order); no quest drill rows; smoke only).
- [ ] #6 Amazon: a fresh-hero live Normal run (no cheats, writes after start 0) reaches the end of Act N; record N and the real clvl (today: no uncheated live run recorded here, U).
- [ ] #7 Amazon: known blockers listed and closed. Current: None known beyond: no drill rows, no live run.
<!-- AC:END -->
