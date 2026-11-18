---
id: TASK-464
title: 'uib1: Pause and resume as player commands'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T04:10:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 265000
card: "uib1"
owns:
  - "internal/bot/playerpause.go"
  - "internal/bot/playerpause_test.go"
  - "internal/watch/pauseapi.go"
  - "internal/watch/pauseapi_test.go"
  - "internal/watch/testdata/pause/"
  - "cmd/d2bot/pausewire.go"
  - "docs/pause.md"
  - "internal/bot/runner.go"
  - "internal/bot/goal.go"
  - "internal/bot/playeraction.go"
  - "internal/watch/funwire.go"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/live.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Pause = go home at the next safe moment, then stop; Resume continues; persisted; GET/POST /api/pause (watch.pause/1). Safety code is untouched (it ranks above player commands). Not verified live; checklist in docs/pause.md.

result: built and tested offline (fake runner, Sorceress and Paladin, route goldens); not verified live.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fake-runner tests, Sorceress and Paladin: pause in the field goes home first and then stops; pause in town stops at once; a hostile in reach holds the walk home; resume continues; a restart stays paused
- [x] #2 A pause never leaves the hero standing in a field (test: no paused state while outside town)
- [x] #3 Route tests: gates, 503 until bound, goldens
- [x] #4 Live checklist in docs/pause.md
<!-- AC:END -->
