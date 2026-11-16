---
id: TASK-552
title: 'survsd: survival slice S-D, bot.zoneModel switches zone choice to the margin controller (default 0)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T16:45:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 291000
card: "survsd"
owns:
  - "internal/survival"
  - "internal/bot"
  - "internal/power"
  - "internal/env"
  - "cmd/d2bot"
  - "docs/survival"
  - "docs/knobs.md"
  - "tools/knobdump"
exclusive: []
contracts: []
result: "bot.zoneModel (0 legacy default, 1 margin model) and bot.zoneDial (default balanced): Plan.Pick asks the survival controller through PickOptions.Model, falls back to the plan tables on any error, dwell 3 min; Andariel floor lowered by a safe boss margin only when data exists (warn never wall); dead merc left dead by the controller; fresh game through the existing guarded startFreshGame; words-only glance gains model:true. Tests on Sorceress and Paladin; docs/survival/zonemodel.md. NOT verified live; the A/B (ab-harness.md) has not run, default stays 0."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the survival design (2026-11-14) slice S-D: the controller must be able to decide the zone behind a knob so the A/B can run. Done when: the knob exists at default 0, the model arm works offline on two builds, abrun.sh preflight passes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 bot.zoneModel and bot.zoneDial registered, default 0 and balanced; Plan.Pick asks the controller when 1, falls back with a log line on error
- [x] #2 boss floor (Andariel) replaced by the margin only when boss margin data exists; merc dead policy; fresh game through the single guarded startFreshGame
- [x] #3 words-only margin in Status/glance; tests on Paladin and Sorceress; docs and knobs updated; abrun.sh preflight passes
<!-- AC:END -->
