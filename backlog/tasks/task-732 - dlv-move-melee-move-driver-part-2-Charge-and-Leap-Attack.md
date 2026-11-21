---
id: TASK-732
title: 'dlv-move-melee: move driver part 2: Charge and Leap Attack'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:44:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 355000
card: "dlv-move-melee"
owns:
  - "internal/combat/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Charge and Leap Attack as gap closers (aim target and through), Leap as an escape. Depends on dlv-move-tele, dlv-press.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Row added to combat.teleport-reposition for a Barbarian and a Paladin
- [ ] #2 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #3 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #4 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
