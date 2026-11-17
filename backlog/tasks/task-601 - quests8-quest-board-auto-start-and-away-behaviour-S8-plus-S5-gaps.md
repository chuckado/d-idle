---
id: TASK-601
title: 'quests8: quest board auto-start and away behaviour (S8) plus S5 gaps'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T08:08:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 305000
card: "quests8"
owns:
  - "internal/questboard"
  - "internal/bot/questboard.go"
  - "internal/bot/questauto.go"
  - "internal/bot/questauto_test.go"
  - "internal/env/tuning_quests.go"
  - "internal/watch/questsapi.go"
  - "internal/watch/questsapi_test.go"
  - "internal/watch/questui_test.go"
  - "internal/watch/web/panels/quest-strip"
  - "internal/watch/web/panels/quest-board"
  - "internal/watch/testdata/questui"
  - "tools/uishots/fixtures_quests.go"
  - "tools/uishots/fixtures_quests_test.go"
  - "cmd/d2bot/questwire.go"
  - "docs/questboard.md"
  - "docs/design/quest-board-2026-11-16.md"
  - "docs/knobs.md"
  - "internal/bot/playerpause.go"
  - "internal/bot/playerreset.go"
  - "internal/watch/testdata/quests"
  - "internal/watch/web/panels/quest-list"
  - "internal/watch/web/panels/needs-you"
  - "internal/watch/web/panels/orders"
  - "tools/uishots.sh"
  - "internal/watch/web/utilities.css"
exclusive: []
contracts: []
result: "S8 built offline behind quests.board: per-quest auto flag and POST /api/quests/<id>/auto {on}, safe-only auto-start with countdown/Cancel, hold after pause or reset, gap and day cap, away detection, auto hand-in only for board-started chunks while away, while-you-were-away log, card and Needs row; S5 gaps closed (step list, levels to go in words, offer Needs row, 360x640 Needs overflow). Gate PASS, fit quest* all sizes incl 360x640. NOT verified live (checklist in docs/questboard.md)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: opt-in auto-start when away (design S8, user decisions 2026-11-16). Done when: merged behind quests.board; live run later.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Auto flag route, executor rule (safe only, countdown, rate limits, away), away log and API field
- [x] #2 S5 gaps: step list, levels to go, offer Needs row, 360x640 Needs overflow
- [x] #3 UI: auto toggle with countdown and Cancel, while-you-were-away card, fit fixtures, node tests
- [x] #4 Tests on two classes with fake clocks; knob 0 identical; docs and live checklist
<!-- AC:END -->
