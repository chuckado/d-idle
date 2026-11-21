---
id: TASK-719
title: 'dlv-rotation: Rotation type and FromProfile adapter, golden parity'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:41:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 342000
card: "dlv-rotation"
owns:
  - "internal/delivery/"
  - "internal/combat/"
  - "internal/agency/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Rotation type and conditions (appendix B s4); FromProfile adapter; Fighter attack order, buff groups and aura ladder read slots; agency lays its plan as a Rotation. Gate: golden parity over the existing simulator tests and scenario corpus (Q11), byte-identical decisions. Owns the golden parity harness reused by later cards. Depends on dlv-data.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Golden parity test over simulator tests and combat scenarios of the corpus: byte-identical decisions
- [ ] #2 Scenario combat.parity-rotation
- [ ] #3 Must-fail unit tests on Paladin and Sorceress
- [ ] #4 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #5 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #6 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
