---
id: TASK-531
title: 'idfast: identify and single-item actions feel instant, identify requests protected from sell'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T02:51:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 284000
card: "idfast"
owns:
  - "internal/invdesk"
  - "internal/watch/web/inventory-actions.js"
  - "internal/watch/web/inventory-actions.css"
  - "internal/watch/testdata/uic3"
  - "tools/uishots/fxviews.go"
  - "docs/inventory-actions.md"
  - "internal/bot/playeraction.go"
  - "internal/bot/playeraction_test.go"
  - "internal/town/chores.go"
  - "cmd/d2bot/adapters.go"
exclusive: []
contracts: []
result: "Diagnosed: lane off by default, 10 s player gap after every attempt (fixed), 4 s page poll, no feedback. UI: optimistic Identifying state, 200 ms poll while pending, 2 s highlight, refusal revert with reason, wait reason on the item, lost/failed toast, Cain progress, no-tome offer, Instant actions: on indicator. Backend: identify request implies keep (sell/tidy/stash skip it), lost requests reported, chore skip reasons logged. V offline (node c3.js, Go tests, Paladin and Sorceress). NOT verified live; coordinator sets invdesk.instant via tuning.json (docs/inventory-actions.md)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

User 2026-11-15: identify too slow; the bot sold items queued for identify.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 diagnosis written (docs/inventory-actions.md)
- [x] #2 UI optimistic state, fast poll, wait reason
- [x] #3 identify request protects the item from sell, lost requests reported
- [x] #4 node and Go tests pass
<!-- AC:END -->
