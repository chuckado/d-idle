---
id: TASK-788
title: 'verbs2: buy, sell, waypoint, portal and pickup input verbs (Q13, ADR 0021)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T19:24:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 361000
card: "verbs2"
owns:
  - "internal/d2dctl/world_verbs.go"
  - "internal/d2dctl/world_verbs_test.go"
  - "internal/d2dctl/fake/fake.go"
  - "internal/input/worldverbs.go"
  - "internal/input/worldfake.go"
  - "internal/d2dinput/worldverbs.go"
  - "internal/d2dinput/controller.go"
  - "internal/d2dinput/controller_test.go"
  - "internal/d2dcombat/rig.go"
  - "internal/d2dtown/adapter.go"
  - "internal/town/shop.go"
  - "internal/town/worldverbs_test.go"
  - "internal/town/worldverbs_unit_test.go"
  - "internal/town/export_world_test.go"
  - "internal/nav/waypoint.go"
  - "internal/nav/wpverb_test.go"
  - "internal/env/tuning_input.go"
  - "docs/knobs.md"
  - "data/input/caps/d2d-fork-input-verbs2.json"
  - "data/input/caps_test.go"
  - "docs/process/input-verbs2.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Remaining Q13 input verbs of ADR 0021. Done when the fork differential and the bot adoption are merged; a live check needs a sandbox image (a later brief).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fork branch didle/input-verbs2-20261121: buy, sell, waypoint, portal, pickup verbs, caps list them, differential 0 differences, refusals named (pushed to the private remote)
- [ ] #2 Bot: d2dctl client methods and fakes, input.planner adoption in town buy/sell, nav waypoint, controller pickup/portal, tests on two classes
- [ ] #3 docs/process/input-verbs2.md; whole-repo gate green
<!-- AC:END -->
