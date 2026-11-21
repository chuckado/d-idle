---
id: TASK-721
title: 'dlv-reads: flag-gated read verbs: minions, skillstate, combat events'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:41:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 344000
card: "dlv-reads"
owns:
  - "internal/d2dctl/"
  - "docs/process/combat-reads.md"
  - "internal/scenario/"
  - "tools/provingground/"
exclusive:
  - "fork-branch:didle/combat-reads"
  - "d2d-lab-binary"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Fork read verbs (Q4, reads under ADR 0022) flag-gated and listed in caps; bot typed clients and scenario assertion kinds of appendix C s3. Hero acts only on reads a player can see (Q5): rows marked visible/hidden, hidden reads test-only. Fork work: same-day push to the private remote, no patch here. Depends on dlv-data.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fork tests in the style of input_verbs_d2d; one scenario per assertion kind that fails when inverted
- [ ] #2 Ticks/s measured with the flag on and off
- [ ] #3 Visible/hidden mark on every read row
- [ ] #4 Foundation gate: golden parity unchanged
- [ ] #5 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #6 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #7 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
