---
id: TASK-595
title: 'survreask: re-ask the zone model after a cost event'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T06:29:00Z'
updated_date: '2026-11-17T10:24:00Z'
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
result: "Cost events (chicken, death, merc death, potion crisis) re-ask the zone model (bot.zoneModel 1 only, default 0 unchanged). Paladin 8-seed A/B PASS (XP 1.005, deaths 1/0, chickens 15/2); holdout FAIL on seed 149 (chickens from the recovery path, not reached by the re-ask); Sorceress inconclusive. Not flipped. Verified on lockstep d2d, not game.exe."
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
- [x] #3 Holdout seeds and Paladin 8-seed A/B rerun with abreport, doc written
<!-- AC:END -->
