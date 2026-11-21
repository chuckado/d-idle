---
id: TASK-728
title: 'dlv-aim-corpse: corpse aims as data, one shared corpse reservation'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:43:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 351000
card: "dlv-aim-corpse"
owns:
  - "internal/combat/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). corpseBest, corpseNearPack as data; one corpse reservation shared by raise, revive, explode, Death Sentry; Find Potion and Find Item as utility rows; removes necro_corpse.go. Depends on dlv-upkeep-minions, dlv-cast-verbs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Scenario combat.corpse-explodes
- [ ] #2 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #3 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #4 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
