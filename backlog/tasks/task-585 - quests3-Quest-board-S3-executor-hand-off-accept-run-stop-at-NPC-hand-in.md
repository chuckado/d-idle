---
id: TASK-585
title: 'quests3: Quest board S3, executor hand-off (accept, run, stop at NPC, hand-in)'
status: In Progress
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
  - "docs/design/quest-board-2026-11-16.md"
  - "docs/knobs.md"
  - "tools/knobdump/main.go"
  - "internal/env/tuning.go"
  - "internal/bot/goal.go"
  - "internal/bot/task.go"
  - "internal/bot/runner.go"
exclusive: []
contracts: []
result: "S3 built offline behind knob quests.board (default 0): questboard.Exec (Offers, Accept, HandIn, State, Next/Report, Cancel, Reset, Pause), Fake, bot adapter boardStep with the hand-in mailbox; tests on Paladin and Sorceress, restart at every step, evidence conflicts, retry/park, one chunk at a time; knob 0 leaves the planner path unchanged. NOT verified live: the Act 1 drill checklist is in docs/questboard.md; multi-chunk quests wait for S7, cmd/d2bot wiring for S4."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: a Start tap must run exactly one chunk and the bot must stop at the NPC so the player taps Hand in (design 2026-11-16, slice S3). Done when: the executor, hand-in action and park semantics are offline-tested behind quests.board and the live drill is written.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Knob quests.board (default 0) and Exec API Offers/Accept/HandIn/State with fakes
- [x] #2 Accept, run, awaiting-handin, HandIn read-back, restart, death, retry/blocked, Reset, one chunk at a time tested on two builds
- [x] #3 Knob 0 leaves the planner path unchanged (regression tests)
- [x] #4 Docs, knob table, design doc S3 done, live drill checklist written (not run)
<!-- AC:END -->
