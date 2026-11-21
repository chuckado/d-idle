---
id: TASK-767
title: 'bsm4: quest board SM exec (Exec on the run machine behind quests.boardSM, latch, persisted park, Retry clears the executor group)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:59:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 343000
card: "bsm4"
owns:
  - "internal/questboard/exec.go"
  - "internal/questboard/exec_sm.go"
  - "internal/questboard/exec_sm_test.go"
  - "internal/bot/questboard.go"
  - "internal/env/tuning_quests.go"
  - "docs/knobs.md"
exclusive: []
contracts: []
result: "Exec runs on the machine when quests.boardSM=1 (default 0, old paths untouched): accept = chain + busy + idempotent, stop cascades, a park is stored and survives a restart, Retry also lifts the task executor group (queued for the loop goroutine), latches keep a chunk done, repair is a load step. Existing questboard suite green with QUESTBOARD_SM=1 and 0. Not verified live; the live scenarios (board.sm.one-head, board.retry.runs-a-step) still need a builder lane."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Card 4 of docs/design/fable-review-2026-11-21/quest-board.md section 5. ChunkFact.Denied is new (exact denial, R3): the bot does not set it yet (U), so R3 is covered offline only, as the design says. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 knob quests.boardSM (default 0): Accept/Cancel/Retry/Reset on the machine, park persists across restart, latch replaces the in-process reached/handed, repair at load not in Next
- [x] #2 the whole existing questboard suite passes with QUESTBOARD_SM=1 (three deliberate changes noted) and with the knob off
- [x] #3 must-fail tests on Paladin and Sorceress; race count=20; bot/watch quest tests green
<!-- AC:END -->
