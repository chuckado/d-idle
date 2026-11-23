---
id: TASK-539
title: >-
  pause: Pause freezes the game (lockstep on/off), safe loop wiring, banner and
  Space
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T03:42:00Z'
updated_date: '2026-11-16T16:27:00Z'
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
  - "internal/bot/playerpause.go"
  - "internal/d2denv/freeze.go"
  - "internal/d2denv/freeze_test.go"
  - "internal/watch/pauseapi.go"
  - "internal/watch/pauseapi_test.go"
  - "internal/watch/pauseui_test.go"
  - "internal/watch/web/pauseui.js"
  - "internal/watch/web/pause.css"
  - "internal/watch/testdata/pause"
  - "cmd/d2bot/glancewire_pause.go"
  - "cmd/d2bot/pausewire.go"
  - "internal/gclock/gclock_test.go"
exclusive: []
contracts: []
result: "Pause freezes the game through the fork's runtime lockstep switch (no new verb); the loop, clocks, health probe, Reset and Resume are wired and tested offline on two classes; menu button, banner, dimmed view and Space are in; fit fixtures added but not run. NOT verified live: checklist in docs/pause.md. Small edits outside owns: bot runner.go loop.go playerreset.go, env tuning_bot.go (bot.pauseCapMin), d2denv env.go, d2dctl/fake, watch glance.go http_view.go shell.js watch.html, cmd/d2bot d2dbackend.go behaviors.go glancewire.go, tools/uishots, docs/knobs.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User request 2026-11-15: a Pause button that freezes the game. Done when the checks pass; live checklist is in docs/pause.md.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Pause freezes the d2d clock (lockstep on) and resume thaws it; no verb means no button
- [x] #2 no watchdog, stall, stuck, chicken or farm clock fires while frozen; fake-clock tests on two builds
- [x] #3 POST /api/pause and glance report frozen; auto-resume cap with a Needs you row
- [x] #4 UI: banner, dimmed view, Space shortcut, 44 px button; fit fixtures
- [x] #5 docs/pause.md with a live checklist (not run live)
<!-- AC:END -->
