---
id: TASK-708
title: 'andariel1: why Andariel is not offered at clvl 26, and the Monastery road nav'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T02:53:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 340000
card: "andariel1"
owns:
  - "internal/questboard/desk.go"
  - "cmd/d2scenario/tap.go"
  - "internal/nav/exit.go"
  - "tools/provingground/scenarios/act1more.json"
  - "internal/d2map/tracker.go"
  - "internal/d2dworld/world.go"
  - "internal/d2dworld/world_test.go"
  - "internal/nav/exit_test.go"
  - "internal/questboard/desk_test.go"
  - "cmd/d2scenario/tap_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: act1more finding F3/F4 (docs/probes/provingground-act1more-2026-11-20.md): Andariel never offered at clvl 26 and the Monastery Gate to Outer Cloister exit not found. Done when: the board says why, the nav waits for a late border, and the scenario runs.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Board row says why Andariel is held (test, two builds)
- [ ] #2 Tapper logs why a chunk is not startable
- [ ] #3 Monastery road nav fixed with a must-fail test, two builds, or shown to be a scenario defect
- [ ] #4 Scenario act1.andariel.slaughter GREEN or the reason asserted, Paladin 11 and Sorceress 3
<!-- AC:END -->
