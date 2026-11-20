---
id: TASK-688
title: 'k1layout: level layout read verb consumer (quests.layoutRead)'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-20T14:49:00Z'
labels:
  - d2d
  - feature
dependencies: []
priority: medium
ordinal: 335000
card: "k1layout"
owns:
  - "docs/process/layout-read.md"
  - "internal/d2dctl"
  - "internal/d2dworld"
  - "cmd/d2bot/objects.go"
  - "cmd/d2bot/objects_layout_test.go"
  - "internal/env/tuning_quests.go"
  - "internal/d2denv"
  - "docs/knobs.md"
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
- [ ] #1 debug levelmap client, fake, d2dworld merge, knob quests.layoutRead default 0
- [ ] #2 must-fail tests with the fake: knob on, off, verb absent
- [ ] #3 fork verb built and differential 0 differences (live-check)
<!-- AC:END -->
