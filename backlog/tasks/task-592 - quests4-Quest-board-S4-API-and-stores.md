---
id: TASK-592
title: 'quests4: Quest board S4, API and stores'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T05:30:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 303000
card: "quests4"
owns:
  - "internal/questboard"
  - "internal/watch/questsapi.go"
  - "internal/watch/questsapi_test.go"
  - "internal/watch/glance.go"
  - "internal/watch/funwire.go"
  - "internal/watch/testdata/quests"
  - "internal/bot/runner.go"
  - "cmd/d2bot/survivalwire.go"
  - "cmd/d2bot/behaviors.go"
  - "cmd/d2bot/glancewire.go"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/d2dbackend.go"
  - "cmd/d2bot/live.go"
  - "cmd/d2bot/questwire.go"
  - "cmd/d2bot/questwire_test.go"
  - "internal/bot/questboard.go"
  - "internal/bot/questboard_test.go"
  - "docs/questboard.md"
  - "docs/design/quest-board-2026-11-16.md"
exclusive: []
contracts: []
result: "S4 built offline behind knob quests.board (default 0): questboard.Desk, GET /api/quests and POST start, handin (202, async), cancel, retry, mode on the answer mux with the shared gate; every reply has ok, code, reason, already; glance quest strip and Needs rows; per-hero plan store with mode; runner builds the board per hero; Assess replaces the legacy words through the read-only margin query. Tests on Paladin and Sorceress with the Fake, goldens, double press, refusals, auth, restart persistence. NOT verified live: the curl checklist is in docs/questboard.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the player needs the quest board over HTTP (design section 6, slice S4). Done when: routes, stores, wiring and docs are merged behind quests.board.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GET/POST /api/quests routes, refusals, idempotency, auth, goldens
- [x] #2 cmd/d2bot constructs the board behind quests.board and wires Assess
- [x] #3 per-hero plan store, mode, restart persistence
- [x] #4 docs and live curl checklist
<!-- AC:END -->
