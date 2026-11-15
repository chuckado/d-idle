---
id: TASK-471
title: 'uib6: Journey blocked row on Glance and Commands'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T04:53:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 267000
card: "uib6"
owns:
  - "cmd/d2bot/glancewire_blocked.go"
  - "cmd/d2bot/glancewire_blocked_test.go"
  - "cmd/d2bot/glancewire.go"
  - "cmd/d2bot/funwire.go"
  - "internal/watch/commands.go"
exclusive: []
contracts: []
result: "Journey blocked row on /api/glance and blocked[] on /api/commands from runner status; tested for two classes; not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: the bot knows why the quest step is refused but the watch screen did not show it outside the Loop row. Done when: goals.Status.JourneyBlocked feeds a warn Needs row on /api/glance ("Journey blocked: <reason>", actions Go now and Hold chapter on /api/campaign/intent for the decision chapter) and blocked[] of /api/commands (also without a Loop; duplicate with the loop row removed in commands.go). Empty reason clears both. Needs the glanceExtra hook of uib5 (merged in).

Live check (not run): reach a state where the quest step is refused (no mercenary yet) and read GET /api/glance and /api/commands.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Test: a status with the reason yields the Needs row and the blocked[] entry; clearing removes both; two classes
- [x] #2 No existing golden changes
<!-- AC:END -->
