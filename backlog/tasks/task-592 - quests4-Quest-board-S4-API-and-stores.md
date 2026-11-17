---
id: TASK-592
title: 'quests4: Quest board S4, API and stores'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T05:30:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 303000
card: "quests4"
owns:
  - "internal/questboard"
  - "internal/watch/questsapi.go"
  - "internal/watch/questsapi_test.go"
  - "internal/watch/glance_quest.go"
  - "cmd/d2bot/questwire.go"
  - "cmd/d2bot/questwire_test.go"
  - "internal/bot/questboard.go"
  - "internal/bot/questboard_test.go"
  - "docs/questboard.md"
  - "docs/design/quest-board-2026-11-16.md"
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
- [ ] #1 GET/POST /api/quests routes, refusals, idempotency, auth, goldens
- [ ] #2 cmd/d2bot constructs the board behind quests.board and wires Assess
- [ ] #3 per-hero plan store, mode, restart persistence
- [ ] #4 docs and live curl checklist
<!-- AC:END -->
