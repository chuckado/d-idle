---
id: TASK-725
title: 'dlv-stance-army: stance behindArmy and army facts for the arbiter'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:42:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 348000
card: "dlv-stance-army"
owns:
  - "internal/combat/"
  - "internal/arbiter/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Stance behindArmy; arbiter facts army below want and upkeep due; retreat when census under stance minimum in a boss fight; bounded by the cast-first watchdog. Depends on dlv-upkeep-minions.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Scenario combat.army-holds extended with a boss row
- [ ] #2 Knob
- [ ] #3 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #4 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #5 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
