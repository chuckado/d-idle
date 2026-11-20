---
id: TASK-681
title: 'cainchain: Search for Cain chain re-entry, stale town portal, determinism'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T02:38:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 333000
card: "cainchain"
owns:
  - "internal/task/act1.go"
  - "internal/task/cainreenter_test.go"
  - "internal/town/trip.go"
  - "internal/town/portalonly_test.go"
  - "internal/nav/route.go"
  - "docs/probes/provingground-cainchain-2026-11-19.md"
  - "internal/town/town.go"
  - "internal/data/db.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 cage step re-enters Tristram through the Stony Field portal
- [ ] #2 BackToField refuses a stale portal for a portal-only level
- [ ] #3 cain.chain green 3 runs Paladin 11, 5, Sorceress 3 and a determinism pair
<!-- AC:END -->
