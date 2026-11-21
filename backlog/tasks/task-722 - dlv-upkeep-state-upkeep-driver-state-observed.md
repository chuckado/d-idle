---
id: TASK-722
title: 'dlv-upkeep-state: upkeep driver, state-observed'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:42:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 345000
card: "dlv-upkeep-state"
owns:
  - "internal/combat/"
  - "internal/data/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Upkeep driver for slots with observe.by state/aura/clock; exclusive groups; in-fight share data defaults per archetype (1 cast in 4, mana reserve 10 percent, Q7); post-fight pass; deletes buffStates, curse skip list in buff pass, skill-named aura knob. Depends on dlv-rotation, dlv-expect.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Scenario combat.buffs-stay-up RED first, then green for Sorceress and Barbarian
- [ ] #2 Knob combat.delivery.upkeep
- [ ] #3 Differential rows for war-cry and Holy Shield states
- [ ] #4 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #5 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #6 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
