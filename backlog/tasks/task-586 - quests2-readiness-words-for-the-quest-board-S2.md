---
id: TASK-586
title: 'quests2: readiness words for the quest board (S2)'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T03:29:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 302000
card: "quests2"
owns:
  - "internal/questboard"
  - "data/questboard"
  - "docs/questboard.md"
  - "docs/design/quest-board-2026-11-16.md"
exclusive: []
contracts: []
result: "S2 built offline: pure Assess (safe/risky/too hard per dial, reason codes, legacyFloor beside it, sentence), thresholds in data/questboard/readiness.json, power adapter; tested on Paladin and Sorceress; no bot behaviour change; NOT verified live (made-up zones, no game)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: quests need a word per offer (design section 3). Done when: Assess and tests are merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Readiness struct and pure Assess: safe/risky/too hard from the survival margin per dial, reason codes, legacyFloor beside it, human sentence
- [x] #2 Thresholds are data (data/questboard/readiness.json), mirror the power dials (test)
- [x] #3 Unit tests on Paladin and Sorceress against the logged outcomes; unknown readiness keeps the quest available
- [x] #4 docs/questboard.md Readiness section; S2 marked done in the design doc
<!-- AC:END -->
