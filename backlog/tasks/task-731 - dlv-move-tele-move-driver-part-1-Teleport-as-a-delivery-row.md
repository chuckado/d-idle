---
id: TASK-731
title: 'dlv-move-tele: move driver part 1: Teleport as a delivery row'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:44:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 354000
card: "dlv-move-tele"
owns:
  - "internal/combat/"
  - "internal/nav/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Navigation asks the driver for hops (range, cost, refusal from the row and skillstate); in-fight reposition out of a surround; escape hop offered to Reflex; removes telekey.go. Driver is the only caller of the cast; nav keeps the route. Depends on dlv-expect, dlv-reads.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Scenario combat.teleport-reposition
- [ ] #2 Existing teleport drill D22 stays green
- [ ] #3 Queue row 16 and A5 verdicts consulted
- [ ] #4 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #5 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #6 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
