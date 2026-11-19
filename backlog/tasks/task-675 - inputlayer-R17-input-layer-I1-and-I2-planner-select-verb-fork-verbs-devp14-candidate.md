---
id: TASK-675
title: >-
  inputlayer: R17 input layer I1 and I2 (planner, select verb, fork verbs,
  devp14 candidate)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-19T18:56:00Z'
updated_date: '2026-11-19T20:11:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 329000
card: "inputlayer"
owns:
  - "internal/input"
  - "data/input"
  - "internal/env/tuning_input.go"
  - "internal/combat/bindleft.go"
  - "internal/d2dcombat/rig.go"
  - "tools/provingground/scenarios/input-scenarios.json"
  - "docs/process/sandbox-devp14.md"
  - "internal/env/env.go"
  - "internal/env/tuning.go"
  - "internal/d2dinput/controller.go"
  - "internal/bot/observe.go"
  - "internal/scenario"
  - "cmd/d2scenario/run.go"
  - "tools/knobdump/main.go"
  - "docs/knobs.md"
  - "docs/process/proving-ground.md"
  - "tools/provingground/scenarios/behaviours.json"
exclusive: []
contracts: []
result: "Offline only, NOT verified live. Bot: internal/input planner (verb, key, verified click), knob input.planner default 0, SetRight/SetLeft migrated, minute line input=verb/key/click, Proving Ground right_skill_is and scenario input.skill-select-verb (pending). Fork branch didle/input-verbs-20261119: cmd select fixed (V on the builder: stays selected), cmd repair, menu choose (both unverified: headless hero never reached a vendor), caps input.*. Image d2dbot:devp14 f5bc525d159c loaded, not switched; the live check and the switch need the owner's go."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: the bot clicks where the fork has verbs (07-input-layer.md, R17 I1 and I2; Q13 is open, so everything is behind the knob input.planner, default 0). Done when: planner and skill select behind the knob, fork select fix built, devp14 loaded not switched.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/input planner chooses verb, key, verified click from a per-backend capability set; knob input.planner default 0 changes nothing
- [x] #2 skill and aura selection migrated behind the knob, path counters in the minute line
- [x] #3 Proving Ground scenario input.skill-select-verb and the right_skill_is assertion
- [x] #4 fork branch didle/input-verbs-20261119: select fix, repair, menu choose, caps, fork test, docs (private fork only)
- [x] #5 image d2dbot:devp14 loaded on the sandbox host, NOT switched (live check, needs the owner's go)
<!-- AC:END -->
