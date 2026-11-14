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
result: "built, NOT verified live (2026-11-14): merc.PlayerRevive (request, button, cooldown), GET /api/merc (watch.merc/1) and POST /api/merc/revive and /cancel, errand trip through the inventory desk's player-command hooks, the existing gold-only reviveStep does the revive. Tests for Sorceress and Paladin, route goldens, 503 until bound; live checklist in docs/merc-revive.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the policy revives only when a town visit happens to come; the player wants a button (C-orders revive_merc). Done when: the three boxes hold.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fake tests for two classes: dead + enough gold -> trip and revive; gold short -> button off with the amount; alive -> off; failure -> reason and cooldown
- [x] #2 Route tests with goldens; 503 until bound
- [x] #3 Live checklist in the doc
<!-- AC:END -->
