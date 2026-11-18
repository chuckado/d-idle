---
id: TASK-651
title: 'survhold2: survival model gold gap, recovery without potions, portal mismatch, holdout 2'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T18:40:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 324000
card: "survhold2"
owns:
  - "internal/survival"
  - "internal/bot/kitgate.go"
  - "internal/bot/kitlevel_test.go"
  - "internal/bot/portalburn_test.go"
  - "internal/bot/zonemodel.go"
  - "internal/bot/goal.go"
  - "tools/survival"
  - "docs/probes/survival-holdout2-2026-11-18.md"
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
- [ ] #1 Gold gap per zone analysed and recorded
- [ ] #2 Recovery holds a broke potion-less hero on the first fields (must-fail test)
- [ ] #3 Return-portal mismatch fixed (test)
- [ ] #4 Holdout 2 A/B per class on the builder, writes after start 0
- [ ] #5 Probe doc with verdict per class and recommendation
<!-- AC:END -->
