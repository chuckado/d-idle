---
id: TASK-548
title: 'charsup: container entrypoint runs the supervisor for Character Select switches'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T16:08:00Z'
labels:
  - container
  - feature
dependencies: []
priority: medium
ordinal: 289000
card: "charsup"
owns:
  - "deploy/d2d-container/entrypoint.sh"
  - "deploy/d2d-container/tester_image_test.go"
  - "internal/charsel"
  - "cmd/d2bot/charwire.go"
  - "cmd/d2bot/charwire_test.go"
  - "docs/charselect.md"
exclusive: []
contracts: []
result: "Built offline, not verified live: tester profile defaults to the supervisor (D2BOT_SUPERVISE), active-char keeps the switched hero across container restarts (atomic, rolled back on a failed swap), entrypoint prefers it only with an existing save. Friends pack needs the next release (image rebuild). Live checklist 8-12 in docs/charselect.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Tester profile runs d2bot supervised (D2BOT_SUPERVISE default 1, 0 turns off), plain bot unchanged (test)
- [x] #2 active-char written atomically at a switch, rolled back when the swap fails; entrypoint prefers it only with an existing save (tests)
- [x] #3 docs/charselect.md: supervisor, crash-loop guard, live checklist, friends pack impact
<!-- AC:END -->
