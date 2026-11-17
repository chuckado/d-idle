---
id: TASK-601
title: 'quests8: quest board auto-start and away behaviour (S8) plus S5 gaps'
status: Proposed
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
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: opt-in auto-start when away (design S8, user decisions 2026-11-16). Done when: merged behind quests.board; live run later.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Auto flag route, executor rule (safe only, countdown, rate limits, away), away log and API field
- [ ] #2 S5 gaps: step list, levels to go, offer Needs row, 360x640 Needs overflow
- [ ] #3 UI: auto toggle with countdown and Cancel, while-you-were-away card, fit fixtures, node tests
- [ ] #4 Tests on two classes with fake clocks; knob 0 identical; docs and live checklist
<!-- AC:END -->
