---
id: TASK-442
title: 'botmerc: Revive mercenary as a player command'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T21:08:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 259000
card: "botmerc"
owns:
  - "internal/merc/playerrevive.go"
  - "internal/merc/playerrevive_test.go"
  - "internal/watch/mercapi.go"
  - "internal/watch/mercapi_test.go"
  - "cmd/d2bot/mercwire.go"
  - "cmd/d2bot/mercplayer.go"
  - "cmd/d2bot/mercplayer_test.go"
  - "docs/merc-revive.md"
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
- [ ] #1 Fake tests for two classes: dead + enough gold -> trip and revive; gold short -> button off with the amount; alive -> off; failure -> reason and cooldown
- [ ] #2 Route tests with goldens; 503 until bound
- [ ] #3 Live checklist in the doc
<!-- AC:END -->
