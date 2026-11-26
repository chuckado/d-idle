---
id: TASK-1003
title: 'whynot1: arbiter decision in explain, Why-not row per chunk'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T17:44:00Z'
updated_date: '2026-11-26T18:33:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 482000
card: "whynot1"
lane: next
owns:
  - "internal/questboard"
  - "internal/arbiter/explained.go"
  - "internal/arbiter/explained_test.go"
  - "internal/bot/explain.go"
  - "internal/bot/explain_test.go"
  - "internal/ops"
  - "cmd/d2bot/opscmd.go"
  - "cmd/d2bot/questwire.go"
  - "internal/watch/web/panels/quest-board"
  - "internal/watch/questui_test.go"
  - "internal/watch/testdata/quests"
  - "docs/operator.md"
exclusive: []
contracts: []
result: "The explain record carries the arbiter decision (intent, family, reason, what it displaced) and every waiting chunk a Why-not row (readiness word, floor, need, wake, admission words, last refusal) on the quest board sheet and in ops quest status --why. Unit tests only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

P7 remainder (Fable review 2026-11-25).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 arbiter decision in the explain record
- [x] #2 Why-not row per chunk on the watch page and in ops quest status --why
<!-- AC:END -->
