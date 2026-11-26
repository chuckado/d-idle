---
id: TASK-956
title: 'outcome1: typed step outcome, stage 1 (quest board refusal)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:41:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 455000
card: "outcome1"
owns:
  - "internal/bot/outcome.go"
  - "internal/bot/outcome_test.go"
  - "docs/outcomes.md"
  - "internal/bot/loop.go"
  - "internal/bot/runner.go"
  - "internal/bot/questboard.go"
  - "internal/env/tuning_bot.go"
exclusive: []
contracts: []
result: "outcome1 2026-11-25, offline only (not verified live): StepOutcome type + needTable, quest board refusal converted behind bot.typedOutcome (default 1), handled in one place (no strike, one INFO line per change); 5 follow-up cards for the remaining paths."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. P2 of the 2026-11-25 logic review, stage 1.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 StepOutcome type (done, progress, refused, failed, interrupted) and the need table
- [x] #2 quest board refusal converted: handled in one place, no strike, one line per change, legacy path behind bot.typedOutcome
- [x] #3 outcome table documented, remaining paths listed as follow-up cards
<!-- AC:END -->
