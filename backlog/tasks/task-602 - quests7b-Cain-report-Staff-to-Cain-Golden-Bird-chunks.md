---
id: TASK-602
title: 'quests7b: Cain report, Staff to Cain, Golden Bird chunks'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T08:34:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 306000
card: "quests7b"
owns:
  - "internal/task/act1.go"
  - "internal/task/act2.go"
  - "internal/task/act3.go"
  - "data/questboard"
  - "docs/questboard.md"
  - "internal/task/world.go"
  - "internal/task/s7b_test.go"
  - "internal/task/testdata"
  - "internal/talk/tables.go"
  - "internal/bot/questboard.go"
  - "internal/bot/questboard_s7b_test.go"
  - "internal/questboard/chunks_test.go"
  - "internal/questboard/state_test.go"
exclusive: []
result: "Cain report, Staff assemble (hand-in to Cain) and Golden Bird chunks are supported through board-only task groups (search_for_cain_board, horadric_staff_board, golden_bird_board); campaign groups unchanged except Cain: Cleared counts as freed. Seals chunk stays unsupported with a spike plan. NOT verified live: every dialogue row and quest-word rule is U, drills listed in docs/questboard.md."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: S7 left four chunks unsupported. Done when: three are supported with tests, the seals chunk has a sharper reason and a spike plan.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Cain report, Staff assemble and Golden Bird chunks supported with tests on two builds
- [x] #2 Terror's End seals kept unsupported with a spike plan
- [x] #3 live drills listed in docs/questboard.md (not run)
<!-- AC:END -->
