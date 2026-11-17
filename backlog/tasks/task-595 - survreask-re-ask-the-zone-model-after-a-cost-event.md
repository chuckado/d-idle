---
id: TASK-595
title: 'survreask: re-ask the zone model after a cost event'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T06:29:00Z'
updated_date: '2026-11-17T06:37:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 304000
card: "survreask"
owns:
  - "internal/bot/zonemodel.go"
  - "internal/bot/zonereask_test.go"
  - "internal/bot/loop.go"
  - "internal/bot/observe.go"
  - "internal/bot/zoneshadow.go"
  - "internal/survival/controller.go"
  - "cmd/d2bot/survivalwire.go"
  - "docs/probes/survival-ab-reask-2026-11-17.md"
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
- [x] #1 Cost event (chicken, death, merc death, potion crisis) makes the next pick re-ask the zone model, rate limited, logged as zone reask
- [x] #2 Tests on two classes: leave after chicken, no flapping, knob off identical, seed-149 pattern leaves after at most 2
- [ ] #3 Holdout seeds and Paladin 8-seed A/B rerun with abreport, doc written
<!-- AC:END -->
