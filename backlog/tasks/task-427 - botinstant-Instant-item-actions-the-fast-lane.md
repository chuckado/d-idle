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
  - "internal/invdesk/testdata/view_schema9_instant_*.json"
  - "internal/bot/loop.go"
  - "internal/bot/runner.go"
  - "internal/bot/playeraction_test.go"
  - "internal/env/tuning_invdesk.go"
  - "docs/knobs.md"
  - "cmd/d2bot/invverbs_test.go"
  - "internal/invdesk/usable_test.go"
exclusive: []
contracts: []
result: "Fast lane built and tested offline on fakes (Paladin and Sorceress, -race): verb-capable requests run in the POST through a serialised channel shared with the bot, else queue as before; own knob invdesk.instant (default 0); closed on cursor item, open panel and safety action (Runner.SafetyBusy); a lost answer after send is final for gold and drop; view schema 9. Appends the section Instant actions to docs/inventory-actions.md only. NOT verified live (live verification list in that section)."
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
