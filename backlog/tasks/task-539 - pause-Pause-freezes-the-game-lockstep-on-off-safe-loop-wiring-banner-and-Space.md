---
id: TASK-539
title: 'pause: Pause freezes the game (lockstep on/off), safe loop wiring, banner and Space'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T03:42:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 284000
card: "pause"
owns:
  - "docs/pause.md"
  - "internal/bot/freeze.go"
  - "internal/bot/freeze_test.go"
  - "internal/watch/pauseapi.go"
  - "internal/watch/web/pauseui.js"
  - "internal/watch/web/pause.css"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User request 2026-11-15: a Pause button that freezes the game. Done when the checks pass; live checklist is in docs/pause.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Pause freezes the d2d clock (lockstep on) and resume thaws it; no verb means no button
- [ ] #2 no watchdog, stall, stuck, chicken or farm clock fires while frozen; fake-clock tests on two builds
- [ ] #3 POST /api/pause and glance report frozen; auto-resume cap with a Needs you row
- [ ] #4 UI: banner, dimmed view, Space shortcut, 44 px button; fit fixtures
- [ ] #5 docs/pause.md with a live checklist (not run live)
<!-- AC:END -->
