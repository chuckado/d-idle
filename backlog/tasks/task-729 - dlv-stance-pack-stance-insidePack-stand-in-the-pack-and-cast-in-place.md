---
id: TASK-729
title: 'dlv-stance-pack: stance insidePack: stand in the pack and cast in place'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:43:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 352000
card: "dlv-stance-pack"
owns:
  - "internal/combat/"
  - "data/capabilities/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Self-centred insidePack stance after one live look (H1, Q8); Blessed Hammer row corrected; novas and War Cry usable as main. Survival model contact archetype must judge before knob 2; owns capabilities.json only after dlv-aim-point merges. Depends on dlv-rotation, dlv-upkeep-state, dlv-aim-point.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Scenario combat.hammer-in-pack
- [ ] #2 Live look H1 recorded or listed as a ready checklist
- [ ] #3 A6 verdict before War Cry
- [ ] #4 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #5 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #6 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
