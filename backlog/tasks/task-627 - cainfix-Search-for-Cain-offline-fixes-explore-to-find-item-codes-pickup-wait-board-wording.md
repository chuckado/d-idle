---
id: TASK-627
title: 'cainfix: Search for Cain offline fixes (explore-to-find, item codes, pickup wait, board wording)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T00:00:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 314000
card: "cainfix"
owns:
  - "docs/probes/search-for-cain-2026-11-17.md"
  - "cmd/d2bot/objects.go"
  - "cmd/d2bot/objects_test.go"
  - "internal/questboard"
  - "internal/task/world.go"
  - "internal/quest/act1.go"
  - "internal/bot/task.go"
  - "internal/watch"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Why: sandbox hero stuck on Search for Cain (docs/probes/search-for-cain-2026-11-17.md). Done when the offline slices merge; live drills CAIN-L0..L2 stay open.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 diagnosis doc merged
- [ ] #2 CAIN-1 exploring is not a failure
- [ ] #3 CAIN-2 quest items by code
- [ ] #4 CAIN-3 pickup waits after tree click
- [ ] #5 CAIN-6 honest chain wording, CAIN-4/5 shown as can't do yet
<!-- AC:END -->
