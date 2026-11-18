---
id: TASK-635
title: 'cain7: walk to the tree room from the room rows'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-18T14:39:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 319000
card: "cain7"
owns:
  - "cmd/d2bot/objects.go"
  - "cmd/d2bot/objects_room_test.go"
  - "internal/d2dctl/parse.go"
  - "internal/d2dctl/rooms_test.go"
  - "internal/d2dworld"
  - "internal/d2map/tracker.go"
  - "internal/nav/nav.go"
  - "internal/quest/dialog.go"
exclusive: []
contracts: []
result: "Offline only, not verified live: room rows parsed, find walks to room 161 centre, fallback sweep kept; also fixes find's preset lookups never running on d2d (Env.World is nil there). Whether the sandbox d2d prints room rows is U."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the Inifuss tree streams in only within about 13 to 18 cells and sits at a random room per seed; coverage exploring ran out of budget. Done when: the bot walks to the room type 161 centre (room row origin + 0.5) from the presets reply and clicks the tree; the sweep stays the fallback.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 room rows parsed (d2dctl, d2dworld, LevelMap.RoomRows)
- [x] #2 find walks to room type 161 centre; sweep only as fallback
- [x] #3 offline tests: L0 fixtures, five rooms, must-fail walk test
<!-- AC:END -->
