---
id: TASK-629
title: 'retirej: Quest board S9 retire the Journey chapters and the charter on quests'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-18T03:16:00Z'
updated_date: '2026-11-18T04:01:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 316000
result: "Offline only, NOT verified live. quests.board default 1; Journey panels/controller/guide/API removed; plan import of old campaign.json (skips, orders mode) with a read-only backup; zone-order charter kept."
card: "retirej"
owns:
  - "internal/campaign"
  - "internal/watch"
  - "cmd/d2bot"
  - "internal/questboard"
  - "internal/bot"
  - "internal/task"
  - "internal/env"
  - "data/onboarding"
  - "tools/uishots"
  - "docs/questboard.md"
  - "docs/campaign.md"
  - "docs/knobs.md"
  - "docs/onboarding.md"
  - "docs/friend-guide.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: user decision 2026-11-17, slice S9 of docs/design/quest-board-2026-11-16.md: the board is the only quest driver. Done when: the Journey (chapters, strip, chip, sheet, controller, guide, /api/campaign) is gone, quests.board defaults to 1, old heroes migrate, zone-order charter pricing untouched.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 quests.board defaults to 1; the planner no longer plays quests with the Journey
- [x] #2 Journey panels, controller, guide, API and charter on quests removed
- [x] #3 State migration for old heroes (read-only fallback, backup)
- [x] #4 Tests on two classes, whole repo green
<!-- AC:END -->
