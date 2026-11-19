---
id: TASK-675
title: 'inputlayer: R17 input layer I1 and I2 (planner, select verb, fork verbs, devp14 candidate)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T18:56:00Z'
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
- [ ] #1 internal/input planner chooses verb, key, verified click from a per-backend capability set; knob input.planner default 0 changes nothing
- [ ] #2 skill and aura selection migrated behind the knob, path counters in the minute line
- [ ] #3 Proving Ground scenario input.skill-select-verb and the right_skill_is assertion
- [ ] #4 fork branch didle/input-verbs-20261119: select fix, repair, menu choose, caps, fork test, docs (private fork only)
- [ ] #5 image d2dbot:devp14 loaded on the sandbox host, NOT switched (live check, needs the owner's go)
<!-- AC:END -->
