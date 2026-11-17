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
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Chunk data (handler, nodes, hand-in step, unsupported) for Acts 1 to 5, any-order pieces
- [ ] #2 Bot adapter runs a chunk of a multi-chunk quest from its handler nodes
- [ ] #3 Tests: every order permutation, unsupported chunk, evidence conflicts, difficulty change, restart
- [ ] #4 Validation test over all act data (acyclic, handler or unsupported, NPCs exist)
- [ ] #5 docs/questboard.md chunk tables and a live drill list per chain (not run)
<!-- AC:END -->
