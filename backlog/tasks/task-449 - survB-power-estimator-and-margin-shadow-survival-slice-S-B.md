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
  - "cmd/d2bot/dangerwire.go"
  - "cmd/d2bot/powerwire_test.go"
  - "internal/d2/stats.go"
  - "internal/d2dsrc/source_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: slice S-B of docs/survival/design-2026-11-14.md (sections 4, 5, 6, 8, 10): the character side of the survival model, shadow only. Done when: the criteria below hold; log and replay only, no behaviour change.

result: "S-B shipped offline, shadow only, NOT verified live and the pass rule NOT run on real logs. internal/power: five archetypes and the resolver as data (golden table for every shipped build, synthetic class test, errors for unknown class or skill), hero terms with measured/derived/default quality, Margin() with best and low estimates, monotonicity tests for Paladin and Sorceress, words only for the UI; zone shadow: line after every planner line (Deps.ZoneShadow); levelsim -shadow replay and the section 9 pass rule (tested on a synthetic log, raw swap7/h4/hero14 logs are not in the repo); d2d adapter now maps stats.defense to stat 31 and gear block to stat 20 (value semantics U). Live checks and S-C blockers: docs/survival/power.md. TASK-217/218 are folded into S-A/S-B and TASK-219 to 223 superseded: the coordinator closes them."
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 env reads: what hero defence, damage, block, resist and pools reads give today, tagged, with a Fake and a live-verification checklist
- [x] #2 archetype profiles and resolver as data, two-build and synthetic-class tests
- [x] #3 power terms with quality flags, Margin(), monotonicity tests
- [x] #4 zone shadow log line next to planner, no behaviour change; levelsim -shadow replay with the section 9 pass rule
<!-- AC:END -->
