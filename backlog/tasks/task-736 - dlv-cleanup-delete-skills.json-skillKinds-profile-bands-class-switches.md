---
id: TASK-736
title: 'dlv-cleanup: delete skills.json, skillKinds, profile bands, class switches'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:45:00Z'
labels:
  - bot
  - chore
dependencies: []
priority: medium
ordinal: 357000
card: "dlv-cleanup"
owns:
  - "internal/combat/"
  - "internal/power/"
  - "tools/ci/"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Design: docs/design/fable-review-2026-11-21/combat-engine-d-cards.md (user-approved 2026-11-21, all Q1-Q11 as recommended). Delete skills.json, skillKinds, profile bands, layouts, classOptions, Sorceress default helpers; CI literal check at zero for internal/combat.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Every driver knob at 2 for the agreed number of green nightly runs
- [ ] #2 CI literal check at zero
- [ ] #3 Two builds tested (Paladin + Sorceress), no class or skill literal added, literal count reported in result (may only fall)
- [ ] #4 Knob combat.delivery.<driver> 0/1/2 where a driver ships; rollback is the knob
- [ ] #5 go test -race -count=20 on touched packages; tools/ci/check.sh, contracts/backlog checks and leakcheck clean
<!-- AC:END -->
