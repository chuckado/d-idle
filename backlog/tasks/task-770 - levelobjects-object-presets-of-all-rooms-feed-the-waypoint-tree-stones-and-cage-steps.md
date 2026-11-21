---
id: TASK-770
title: 'levelobjects: object presets of all rooms feed the waypoint, tree, stones and cage steps'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T06:32:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 358000
card: "levelobjects"
owns:
  - "internal/d2dctl/levelmonsters.go"
  - "internal/d2dctl/objpresets_test.go"
  - "internal/d2dworld/"
  - "tools/provingground/scenarios/cain.json"
exclusive: []
contracts: []
result: "Done on the builder, NOT verified in the live sandbox. Fork debug levelmonsters gained objpresets (every object preset unit of ALL rooms of the level, built or not; fork branch didle/level-objects-20261121). d2dctl.LevelMonsters.ObjPresets and d2dworld.mergeObjectPresets (knob quests.layoutRead) add them to the level presets; must-fail tests on a Paladin and a Sorceress arrival. act1.wp.dark-wood and .layout are real scenarios and GREEN on the builder (Paladin 11, Sorceress 3, CHEATED setup, mechanics only). Finding: the five Tristram Cairn Stones are not DS1 preset units (none listed); the cage (class 26) and the Dark Wood Tree (class 30) are."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the hero explored the Dark Wood without finding the waypoint because the levelmap lists only built rooms. Done when: objpresets name the waypoint at arrival and act1.wp.dark-wood is green.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fork verb levelmonsters objpresets lists the object presets of all rooms (fork branch didle/level-objects-20261121)
- [x] #2 d2dctl field and d2dworld merge with must-fail tests on Paladin and Sorceress arrivals
- [x] #3 act1.wp.dark-wood scenario run on the builder with the new binary
<!-- AC:END -->
