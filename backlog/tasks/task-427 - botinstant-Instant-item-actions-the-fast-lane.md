---
id: TASK-427
title: 'botinstant: Instant item actions: the fast lane'
status: In Progress
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
  - "internal/invdesk/identify.go"
  - "internal/invdesk/stash.go"
  - "internal/invdesk/unequip.go"
  - "internal/invdesk/testdata/view_schema8_instant_*.json"
exclusive: []
contracts: []
result: "Fast lane built and tested offline on fakes (Paladin and Sorceress, -race): verb-capable requests run in the POST through a serialised channel shared with the bot, else queue as before; view schema 8. NOT verified live (checklist in docs/inventory-actions.md). Needs Fable review before merge: not queued."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Why: a verb-capable item action should not wait for a bot errand and a safe moment (watch UI redesign B8, answers Q14 to Q16). Done when: the boxes below hold.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Fake-backend tests, two classes: each verb-capable kind completes within the POST; each refusal code comes back as the request's reason; with the knob off, no caps, a named run, a busy cursor or a safety action the request is queued as before
- [x] #2 Stash and take: at the chest instant; in the field `wait: town`, then done on arrival; cancel while waiting
- [x] #3 A race test: the bot's own verb use and a player's instant request never interleave on the channel (`-race`)
- [x] #4 Old inventory goldens unchanged; new view goldens with the `instant` block
- [x] #5 The doc section says what stays an errand and why, and carries a live checklist
<!-- AC:END -->
