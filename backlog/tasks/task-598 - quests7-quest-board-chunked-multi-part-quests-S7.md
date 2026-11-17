---
id: TASK-598
title: 'quests7: quest board chunked multi-part quests S7'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T07:31:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 305000
card: "quests7"
owns:
  - "data/questboard"
  - "internal/questboard"
  - "internal/bot/questboard.go"
  - "internal/bot/questboard_test.go"
  - "docs/questboard.md"
  - "docs/design/quest-board-2026-11-16.md"
  - "internal/watch/testdata/quests"
exclusive: []
result: "Chunk data for Acts 1 to 5 (handler, nodes, handInStep, unsupported), any-order pieces, bot runs chunks of multi-chunk quests from handler steps; Cain report, Staff assemble, Golden Bird, Terror's End seals are unsupported with reasons. Offline only, NOT verified live (drill list in docs/questboard.md)."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Chunk data (handler, nodes, hand-in step, unsupported) for Acts 1 to 5, any-order pieces
- [x] #2 Bot adapter runs a chunk of a multi-chunk quest from its handler nodes
- [x] #3 Tests: every order permutation, unsupported chunk, evidence conflicts, difficulty change, restart
- [x] #4 Validation test over all act data (acyclic, handler or unsupported, NPCs exist)
- [x] #5 docs/questboard.md chunk tables and a live drill list per chain (not run)
<!-- AC:END -->
