---
id: TASK-738
title: 'pgtaps: scripted taps and API calls on the virtual clock inside the bot'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-21T04:46:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 344000
card: "pgtaps"
owns:
  - "cmd/d2bot/ladder.go"
  - "cmd/d2scenario/tap.go"
  - "internal/scenario/taps.go"
  - "cmd/d2scenario/run.go"
  - "cmd/d2scenario/tap_test.go"
  - "internal/scenario/result.go"
exclusive: []
contracts: []
result: "STEP 1 of 2 done (group 1): the board tapper polls on a grid of the bot VIRTUAL clock (GET /api/thoughts now; 30 virtual s grid, 250 ms wall poll, wall fallback after 90 s without a reading, logged) instead of 20 wall s plus 5 wall s; Result.taps_at records each tap virtual second. Default behaviour of every board-taps scenario changes: the Start tap now lands on the first grid point after the offer. API calls (trigger.api After) stay wall-clock (H2 open). Scripted taps inside the bot (steps 2) NOT done. Offline: fake board tests, -race -count=20. U live: that /api/thoughts now advances under the lab d2d; the journal pair card (pgjournal) measures it."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design section 2.8 step 3, 7 card 4. Medium risk: touches the bot answer path. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 --scripted-taps file with rules evaluated on the stepping goroutine
- [ ] #2 HTTP tapper stays for sandbox e2e, those runs are never stable
- [ ] #3 two-run pair of one board scenario has equal D1 and equal Start tick
- [x] #4 STEP 1: board tapper polls on virtual time (grid), taps_at recorded, wall fallback, tests
<!-- AC:END -->
