---
id: TASK-510
title: 'botreset: Reset control (soft leave-and-rejoin, hard restart, cockpit button)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T16:59:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 278000
card: "botreset"
owns:
  - "internal/bot/playerreset.go"
  - "internal/bot/playerreset_test.go"
  - "internal/watch/resetapi.go"
  - "internal/watch/resetapi_test.go"
  - "internal/watch/resetui_test.go"
  - "internal/watch/web/resetui.js"
  - "cmd/d2bot/resetwire.go"
  - "cmd/d2bot/resetwire_test.go"
  - "cmd/d2bot/glancewire_reset.go"
  - "docs/reset.md"
  - "internal/watch/testdata/reset"
exclusive: []
contracts: []
result: "Reset built and tested offline, NOT verified live: GET/POST /api/reset (watch.reset/1, goldens, refusals measured_run/hard_off, idempotent double press), soft level through the runner Cycler new game, hard level = process exit with exitStorm like the watchdog, state file kept across the restart, player_reset why code and a Resetting band word, cockpit confirm sheet + strip (menu, Orders More, Needs you Reset chip when stuck), fit fixtures reset-* incl. 412x780 and 360x640. The soft level has one fresh-game call site to swap for startFreshGame when agent/farmloop merges. Live checklist in docs/reset.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GET/POST /api/reset with goldens, refusals, double press
- [x] #2 soft and hard levels, state kept across a reset
- [x] #3 cockpit confirm sheet, strip, menu, More item, Needs you chip
- [x] #4 fit fixtures at 8 sizes and the real phone sizes
- [x] #5 docs/reset.md with the live checklist
<!-- AC:END -->
