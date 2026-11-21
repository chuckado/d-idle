---
id: TASK-724
title: 'dlv-mark: mark driver: curses and taunts as data'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:42:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 347000
card: "dlv-mark"
owns:
  - "internal/combat/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Target choice by when, area cover from row radius, state observed on targets, one mark group per target; ten curses as data; Taunt, Howl, Battle Cry, Cloak of Shadows as rows. On d2d the put rule deviates: scenario floor is a d2d number, stated in the result. Depends on dlv-upkeep-state.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Scenario combat.curse-covers-pack
- [ ] #2 No uptime claim before fidelity A2 verdict
- [ ] #3 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #4 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #5 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
