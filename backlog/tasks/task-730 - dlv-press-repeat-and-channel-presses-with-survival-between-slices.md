---
id: TASK-730
title: 'dlv-press: repeat and channel presses with survival between slices'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:44:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 353000
card: "dlv-press"
owns:
  - "internal/combat/"
  - "internal/d2dcombat/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). repeat and channel presses in slices on the verb path; commit gate; Zeal, Strafe, Fend, Inferno, Arctic Blast rows. Explicit release verb and expectation timeout. Depends on dlv-cast-verbs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Scenario combat.hold-interruptible
- [ ] #2 A1 attack-sequence rate known
- [ ] #3 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #4 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #5 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
