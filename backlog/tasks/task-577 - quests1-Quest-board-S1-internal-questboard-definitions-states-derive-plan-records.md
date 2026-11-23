---
id: TASK-577
title: 'quests1: Quest board S1, internal/questboard definitions, states, derive, plan records'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T02:30:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 301000
card: "quests1"
owns:
  - "internal/questboard"
  - "data/questboard"
  - "docs/questboard.md"
  - "docs/design/quest-board-2026-11-16.md"
exclusive: []
contracts: []
result: "S1 built offline: internal/questboard (schema, loader, validation, nine states, derive, plan store) and Act 1+2 data in data/questboard; not wired until S3; NOT verified live (no game used)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the quest board needs its model before the executor (S3) and UI. Done when: definitions, states, derive and plan store are offline-tested on two builds and nothing in the bot changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Schema, loader and validation (unknown class/act/quest is an error naming it)
- [x] #2 Nine states with evidence tier per transition; derive-from-evidence, derived twice gives the same result
- [x] #3 Plan-record store (atomic, backup, per hero)
- [x] #4 Act 1 and Act 2 data merged from hints, givers, execution.md with tags
- [x] #5 Table-driven tests on Paladin and Sorceress; design doc marks S1 done
<!-- AC:END -->
