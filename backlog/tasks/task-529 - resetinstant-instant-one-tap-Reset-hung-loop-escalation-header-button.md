---
id: TASK-529
title: 'resetinstant: instant one-tap Reset, hung-loop escalation, header button'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:30:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 283000
card: "resetinstant"
owns:
  - "internal/bot/playerreset.go"
  - "internal/bot/playerreset_test.go"
  - "cmd/d2bot/resetwire.go"
  - "cmd/d2bot/resetwire_test.go"
  - "internal/watch/resetapi.go"
  - "internal/watch/resetapi_test.go"
  - "internal/watch/resetui_test.go"
  - "internal/watch/web/resetui.js"
  - "internal/watch/web/reset.css"
  - "internal/watch/testdata/reset"
  - "docs/reset.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Break-glass Reset must be instant (user 2026-11-15). Done when the criteria are ticked.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 soft reset never waits (wait opt-in), starts on the next tick
- [ ] #2 3 s unacknowledged soft reset escalates to hard
- [ ] #3 POST /api/reset answers 202; goldens updated
- [ ] #4 one-tap header Reset button, no confirm for soft, 3 s cooldown
- [ ] #5 fit fixtures for the header at all sizes viewed
- [ ] #6 docs/reset.md updated
<!-- AC:END -->
