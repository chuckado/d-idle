---
id: TASK-630
title: 'flakyfix: deterministic load-sensitive tests'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-18T06:15:00Z'
updated_date: '2026-11-18T07:30:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 317000
card: "flakyfix"
owns:
  - "internal/bot/"
  - "internal/server/"
  - "internal/watch/"
  - "docs/process/flaky-tests.md"
exclusive: []
contracts: []
result: "TestQuestBoardStartRunsUnderAFarmGoal was a real data race: the board world was built on the API goroutine (Accept) while the loop mutated loop state; now built under stepMu / on the loop. TestRingBounded waited 100 ms for the ring goroutine; other 2-3 s positive waits raised to 30 s; a rate test made count-based. Not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: load-sensitive tests failed merge gates. Done when: the race and the short waits are fixed and documented.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 QuestBoardStartRunsUnderAFarmGoal race fixed, -race x300
- [x] #2 other load-sensitive waits made deterministic
- [x] #3 pattern documented
<!-- AC:END -->
