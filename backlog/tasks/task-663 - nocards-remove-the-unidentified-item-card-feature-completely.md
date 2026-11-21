---
id: TASK-663
title: 'nocards: remove the unidentified item card feature completely'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T15:11:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 327000
card: "nocards"
owns:
  - "data/idreveal"
  - "internal/watch/unlockidreveal.go"
  - "internal/watch/unlockidreveal_test.go"
  - "internal/watch/unlockui_quiet_test.go"
  - "internal/watch/testdata"
  - "docs/unlock-ladder.md"
  - "docs/vault.md"
  - "internal/watch/web"
  - "internal/watch/funwire.go"
  - "internal/watch/nocards_test.go"
  - "data/unlocks"
  - "docs/unlock-idreveal.md"
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
- [x] #1 No card, sheet, tab, route or ladder row for unidentified items
- [x] #2 Test: 50 unidentified drops emit no card (server and stub DOM); presets have no Cards tab
<!-- AC:END -->
