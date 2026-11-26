---
id: TASK-997
title: 'arbfam: arbiter families chores claim goldfarm recovery quest level behind bot.arbiterFamilies'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T17:39:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 482000
card: "arbfam"
lane: next
owns:
  - "internal/arbiter"
  - "internal/bot/arbiter_control.go"
  - "internal/bot/arbiter_control_test.go"
  - "internal/bot/shadow.go"
  - "internal/bot/goal.go"
  - "tools/provingground/arbiter-outcomes"
  - "docs/arbiter.md"
  - "data/darklaunch.json"
  - "docs/knobs.md"
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
- [x] #1 chores and claim family hooks defer to the arbiter
- [x] #2 goldfarm and recovery family hooks defer
- [ ] #3 quest family defers
- [ ] #4 level family defers
- [ ] #5 outcome rows and verdicts per family, dark-launch rows
<!-- AC:END -->
