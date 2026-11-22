---
id: TASK-817
title: 'necro5: minion census from the engine read behind combat.minionRead (default 0)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T19:52:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 373000
card: "necro5"
owns:
  - "internal/combat/minionread.go"
  - "internal/combat/necro.go"
  - "internal/combat/fighter.go"
  - "internal/combat/necro_data_test.go"
  - "internal/d2dcombat/rig.go"
  - "internal/env/tuning_combat.go"
  - "docs/knobs.md"
exclusive: []
contracts: []
result: "Fighter.MinionRead + census() behind combat.minionRead (default 0); offline tests only, not verified live; herdr root disk hit 100% on 11-22 (go clean -cache freed 21G), builds run on the worker host."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Knob combat.minionRead 0/1, default 0; rollback is the knob
- [x] #2 Necromancer census from debug minions maps kinds by profile data; failed read falls back; Paladin never consults it
- [ ] #3 Live drill with the fork flag --read-combat (lab night window)
<!-- AC:END -->
