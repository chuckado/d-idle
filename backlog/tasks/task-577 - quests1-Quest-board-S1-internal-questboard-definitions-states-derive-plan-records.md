---
id: TASK-577
title: 'quests1: Quest board S1, internal/questboard definitions, states, derive, plan records'
status: Proposed
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
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Schema, loader and validation (unknown class/act/quest is an error naming it)
- [ ] #2 Nine states with evidence tier per transition; derive-from-evidence, derived twice gives the same result
- [ ] #3 Plan-record store (atomic, backup, per hero)
- [ ] #4 Act 1 and Act 2 data merged from hints, givers, execution.md with tags
- [ ] #5 Table-driven tests on Paladin and Sorceress; design doc marks S1 done
<!-- AC:END -->
