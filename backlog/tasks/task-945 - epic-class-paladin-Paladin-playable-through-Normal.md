---
id: TASK-945
title: 'epic-class-paladin: Paladin playable through Normal'
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
ordinal: 446000
card: "epic-class-paladin"
owns:
  - "docs/epics/epic-class-paladin.md"
exclusive: []
contracts: []
lane: later
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Class epic for the Paladin. Builds in internal/plandata: build-paladin-normal30.json, build-paladin-normal30-v2.json, build-paladin-zealot-lite.json, build-paladin_hammerdin_guide.json, build-paladin_hammerdin_endgame90.json, build-paladin_hammerdin_zeal_endgame90.json; unlock table unlocks-paladin.json. Docs: docs/classes/paladin.md, docs/classes/paladin-probe.md, docs/classes/smoke-paladin.md. What exists today: the live hero class and the class of nearly every matrix row (Paladin seed 11, stand-ins clvl 19 to 30, CHEATED). Blockers: Smith stand click RED, Tools reward word 0x204e (drill.tools.merc17), Andariel margin at clvl 21, Act 2 flaky fights, Duriel hover scan, Acts 3 to 5 RED. Living status: docs/epics/epic-class-paladin.md. Done when: all acceptance criteria are ticked with evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Paladin: class traits, plan table (internal/plandata) and combat profile (internal/combat/classdata) complete and data-only, no class literals in Go; a second-build test passes (build flexibility rule).
- [ ] #2 Paladin: leveling plan clvl 1 to 30+ per the XP table with survival gates (plandata build-*normal30.json), drilled to clvl 30 in the lab.
- [ ] #3 Paladin: stat and skill allocation plan locked and tested (autoalloc, capability table), spends every point.
- [ ] #4 Paladin: kit and economy profile (belt, potions, merc) for the class, no death spiral in a 3 h run.
- [ ] #5 Paladin: quest drills GREEN as Paladin for every act (drill matrix rows with -class Paladin; today: the live hero class and the class of nearly every matrix row (Paladin seed 11, stand-ins clvl 19 to 30, CHEATED)).
- [ ] #6 Paladin: a fresh-hero live Normal run (no cheats, writes after start 0) reaches the end of Act N; record N and the real clvl (today: no uncheated live run recorded here, U).
- [ ] #7 Paladin: known blockers listed and closed. Current: Smith stand click RED, Tools reward word 0x204e (drill.tools.merc17), Andariel margin at clvl 21, Act 2 flaky fights, Duriel hover scan, Acts 3 to 5 RED.
<!-- AC:END -->
