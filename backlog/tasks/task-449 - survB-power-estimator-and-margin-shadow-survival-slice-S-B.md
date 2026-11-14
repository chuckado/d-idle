---
id: TASK-449
title: 'survB: power estimator and margin, shadow (survival slice S-B)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T22:08:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 260000
card: "survB"
owns:
  - "internal/power/"
  - "docs/survival/power.md"
  - "tools/levelsim/shadow.go"
  - "tools/levelsim/shadow_test.go"
  - "tools/levelsim/cmd/levelsim/main.go"
  - "cmd/d2bot/powerwire.go"
  - "internal/bot/zoneshadow.go"
  - "internal/bot/runner.go"
  - "internal/bot/goal.go"
  - "internal/d2dsrc/player.go"
  - "internal/d2dsrc/gearstats.go"
  - "internal/d2dsrc/gearstats_test.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/live.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 env reads: what hero defence, damage, block, resist and pools reads give today, tagged, with a Fake and a live-verification checklist
- [ ] #2 archetype profiles and resolver as data, two-build and synthetic-class tests
- [ ] #3 power terms with quality flags, Margin(), monotonicity tests
- [ ] #4 zone shadow log line next to planner, no behaviour change; levelsim -shadow replay with the section 9 pass rule
<!-- AC:END -->
