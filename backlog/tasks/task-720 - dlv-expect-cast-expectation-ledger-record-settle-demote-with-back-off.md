---
id: TASK-720
title: 'dlv-expect: cast expectation ledger: record, settle, demote with back-off'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:41:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 343000
card: "dlv-expect"
owns:
  - "internal/combat/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Expectation ledger (main doc 3.4): record per cast, settle from later snapshots, per (skill, reason) counters, demotion with back-off, telemetry line; folds in capSeen, curse tries, leftFail, buff cut. Starts log-only (knob 1); demotion is knob 2. Gate: golden parity over the existing simulator tests and scenario corpus (Q11), byte-identical decisions.  Depends on dlv-rotation.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 L1 tests with a refusing simulator, Paladin and Sorceress, must fail without the ledger
- [ ] #2 Scenario combat.dead-skill-falls-back (pending until dlv-reads, reason stated)
- [ ] #3 Golden parity with knob 0 and 1 byte-identical
- [ ] #4 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #5 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #6 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
