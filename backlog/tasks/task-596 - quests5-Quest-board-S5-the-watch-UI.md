---
id: TASK-596
title: 'quests5: Quest board S5, the watch UI'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T06:32:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 304000
card: "quests5"
owns:
  - "internal/watch/web/panels/quest-strip"
  - "internal/watch/web/panels/quest-board"
  - "internal/watch/web/panels/quest-list"
  - "internal/watch/web/panels/video"
  - "internal/watch/web/panels/needs-you"
  - "internal/watch/web/panels/journey-strip"
  - "internal/watch/web/shell.js"
  - "internal/watch/web/shell.json"
  - "internal/watch/web/layout.json"
  - "internal/watch/web/layouts/cockpit.json"
  - "internal/watch/glance.go"
  - "cmd/d2bot/questwire.go"
  - "internal/watch/questui_test.go"
  - "internal/watch/testdata/questui"
  - "tools/uishots"
  - "tools/uishots.sh"
  - "docs/questboard.md"
  - "docs/ui-quests.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Slice S5 of docs/design/quest-board-2026-11-16.md. Done when the pictures at the supported sizes fit and the node tests pass; NOT verified live.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 quest-strip tile in the strip slot (offer, running, awaiting, handing in, blocked, failed) behind glance questBoard
- [ ] #2 quest-board sheet (drawer) and the phone Quests tab: acts, readiness words, parts, auto stub, history, mode switch, empty states
- [ ] #3 Hand in moment read back from GET, payoff toast, real refusal reasons
- [ ] #4 video card quest caption and Needs-you event rows
- [ ] #5 node tests for the state to words mapping and buttons with a fake API
- [ ] #6 fit pages at all sizes and phone, must-fail fixtures, pictures viewed
<!-- AC:END -->
