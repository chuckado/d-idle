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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 bot.zoneModel and bot.zoneDial registered, default 0 and balanced; Plan.Pick asks the controller when 1, falls back with a log line on error
- [ ] #2 boss floor (Andariel) replaced by the margin only when boss margin data exists; merc dead policy; fresh game through the single guarded startFreshGame
- [ ] #3 words-only margin in Status/glance; tests on Paladin and Sorceress; docs and knobs updated; abrun.sh preflight passes
<!-- AC:END -->
