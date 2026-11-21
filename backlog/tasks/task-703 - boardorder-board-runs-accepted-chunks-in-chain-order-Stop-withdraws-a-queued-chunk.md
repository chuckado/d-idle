---
id: TASK-703
title: 'boardorder: board runs accepted chunks in chain order, Stop withdraws a queued chunk'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T00:25:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 338000
card: "boardorder"
result: "Offline fix: board orders accepted chunks by chain order, queued ones wait; Stop withdraws a queued (Locked) record; trigger.api added; unit tests on Paladin+Sorceress green, -race x20 green. Scenarios board.two-accepted-chunks and board.stop-parked-chunk NOT run on the builder (main lane busy). Not verified live."
owns:
  - "internal/questboard"
  - "cmd/d2scenario"
  - "internal/scenario"
  - "tools/provingground/scenarios/cain.json"
  - "internal/watch/web/panels/quest-board"
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
- [x] #1 two accepted chunks run in chain order, queued one waits
- [x] #2 Stop on a queued chunk withdraws it
<!-- AC:END -->
