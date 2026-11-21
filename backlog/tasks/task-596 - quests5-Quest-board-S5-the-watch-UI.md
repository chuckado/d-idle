---
id: TASK-596
title: 'quests5: Quest board S5, the watch UI'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T06:32:00Z'
updated_date: '2026-11-17T08:22:00Z'
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
result: "S5 built offline behind quests.board: quest-strip tile (offer, running, waiting, handing in, confirming, failed, stuck, training, idle), quest-board drawer and phone Quests tab (acts, readiness words, parts, auto stub, history, mode switch, empty state naming the knob), Hand in read back from GET with payoff toast and the API real refusal reasons, video card quest caption, Needs-you event rows, glance questBoard flag retiring the Journey tile only while the board is on. Node tests with a fake API on both builds; fit run quest* at 8 sizes plus phones clean, four must-fail defects. Gaps: no step list in the API (words only), mode and auto are stubs; a Needs row on a 360x640 phone pushes Orders 9 px over (older, not this card). NOT verified live; a look by the user is open."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Slice S5 of docs/design/quest-board-2026-11-16.md. Done when the pictures at the supported sizes fit and the node tests pass; NOT verified live.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 quest-strip tile in the strip slot (offer, running, awaiting, handing in, blocked, failed) behind glance questBoard
- [x] #2 quest-board sheet (drawer) and the phone Quests tab: acts, readiness words, parts, auto stub, history, mode switch, empty states
- [x] #3 Hand in moment read back from GET, payoff toast, real refusal reasons
- [x] #4 video card quest caption and Needs-you event rows
- [x] #5 node tests for the state to words mapping and buttons with a fake API
- [x] #6 fit pages at all sizes and phone, must-fail fixtures, pictures viewed
<!-- AC:END -->
