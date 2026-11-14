---
id: TASK-452
title: 'survC: zone ledger, controller shadow, merc revive value (S-C)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-14T22:31:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 261000
card: "survC"
owns:
  - "internal/survival/"
  - "internal/watch/survivalapi.go"
  - "cmd/d2bot/survivalwire.go"
  - "docs/survival/ledger.md"
  - "internal/danger/gapexport.go"
  - "internal/watch/survivalapi_test.go"
  - "internal/bot/zoneshadow.go"
  - "internal/bot/zoneshadow_test.go"
  - "cmd/d2bot/powerwire.go"
  - "cmd/d2bot/powerwire_test.go"
exclusive: []
contracts: []
result: "Offline only, NOT verified live and NOT replayed on real logs (not in the repo). internal/survival: ledger (versioned ledger.json, decay, k_zone, k_hero, hard events, step-down, left-for-cost lock), Choose (hysteresis, depleted, fresh game, hole), Revive, Economy, words-only Glance; hooks ZoneChoice/ZoneObserve in bot and cmd/d2bot (small edits in runner.go, observe.go, loop.go, areametrics.go, senseshistory.go, d2dbackend.go, live.go, ladder.go, paths.go, watch hub.go and http_view.go); GET /api/survival. Shadow only. Tests on Paladin and Sorceress. TASK-217 to 223 closed in this branch. Docs: docs/survival/ledger.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the margin controller needs observed cost per zone before it may choose zones (S-D). Done when: the ledger and the shadow choice run in log and API only, with tests on two builds.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Per-zone ledger persisted in its own versioned file (old file loads, newer refused), k_zone, k_hero, hysteresis tested on a Paladin and a Sorceress
- [x] #2 Controller shadow choice logged as one 'zone choice:' line next to the planner and zone shadow lines, no behaviour change
- [x] #3 Merc revive value and fresh-game option tested
- [x] #4 Words-only survival glance block served read-only, no numbers
- [x] #5 Real-log replay steps written; TASK-217..223 closed as folded or superseded
<!-- AC:END -->
