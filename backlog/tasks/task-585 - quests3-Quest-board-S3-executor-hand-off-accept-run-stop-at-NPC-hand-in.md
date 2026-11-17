---
id: TASK-585
title: 'quests3: Quest board S3, executor hand-off (accept, run, stop at NPC, hand-in)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T03:28:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 302000
card: "quests3"
owns:
  - "internal/questboard"
  - "internal/bot/questboard.go"
  - "internal/env/tuning_quests.go"
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
- [ ] #1 Knob quests.board (default 0) and Exec API Offers/Accept/HandIn/State with fakes
- [ ] #2 Accept, run, awaiting-handin, HandIn read-back, restart, death, retry/blocked, Reset, one chunk at a time tested on two builds
- [ ] #3 Knob 0 leaves the planner path unchanged (regression tests)
- [ ] #4 Docs, knob table, design doc S3 done, live drill checklist written (not run)
<!-- AC:END -->
