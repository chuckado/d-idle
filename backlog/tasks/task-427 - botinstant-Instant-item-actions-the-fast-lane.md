---
id: TASK-427
title: 'botinstant: Instant item actions: the fast lane'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T15:51:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 255000
card: "botinstant"
owns:
  - "internal/invdesk/instant.go"
  - "internal/invdesk/instant_test.go"
  - "internal/invdesk/desk.go"
  - "cmd/d2bot/invverbs.go"
  - "docs/inventory-actions.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Fake-backend tests, two classes: each verb-capable kind completes within the POST; each refusal code comes back as the request's reason; with the knob off, no caps, a named run, a busy cursor or a safety action the request is queued as before
- [ ] #2 Stash and take: at the chest instant; in the field `wait: town`, then done on arrival; cancel while waiting
- [ ] #3 A race test: the bot's own verb use and a player's instant request never interleave on the channel (`-race`)
- [ ] #4 Old inventory goldens unchanged; new view goldens with the `instant` block
- [ ] #5 The doc section says what stays an errand and why, and carries a live checklist
<!-- AC:END -->
