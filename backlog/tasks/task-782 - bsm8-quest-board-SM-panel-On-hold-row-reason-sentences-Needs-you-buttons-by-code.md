---
id: TASK-782
title: 'bsm8: quest board SM panel (On hold row, reason sentences, Needs-you buttons by code)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T18:04:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 361000
card: "bsm8"
owns:
  - "internal/watch/web/panels/quest-board/panel.js"
  - "internal/watch/testdata/questui/quest_ui.mjs"
  - "internal/questboard/desk.go"
  - "internal/questboard/desk_reason_test.go"
exclusive: []
contracts: []
result: "The quest-board panel shows On hold with the hold sentence for a held run (instead of Running with a step that does not move), a parked run says its park sentence, part rows prefer the reason sentence, and the Needs-you buttons follow the reason code (hold_timeout: Keep waiting and Stop; prereq_missing, quest_locked, unsupported: Stop only; game_failed: none). Node UI test and Go test on Paladin and Sorceress; not looked at on a live page (no screenshot taken)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 stripModel shows On hold with the reason sentence for a held run; parked runs say the park sentence
- [x] #2 Needs-you buttons follow the reason code (Keep waiting, Stop only, none)
<!-- AC:END -->
