---
id: TASK-914
title: 'toolstrade40: stand click reach, healing floor vs belt, no-progress park'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T03:17:00Z'
updated_date: '2026-11-24T03:57:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 419000
card: "toolstrade40"
owns:
  - "cmd/d2bot/objects.go"
  - "cmd/d2bot/objects_reach_test.go"
  - "internal/task/executor.go"
  - "internal/task/noprogress_test.go"
  - "internal/resource/beltspec.go"
  - "internal/bot/goal.go"
  - "internal/bot/healcap_test.go"
  - "internal/bot/questboard.go"
  - "internal/bot/questboard_goal_test.go"
exclusive: []
contracts: []
result: "Offline only, not verified live: stand click needs reach (3 walks, then danger, no clicks wasted); town healing floor capped by HealingCapacity of the belt plan; executor MaxRepeats (12) parks a group with no progress: <step>; a chunk queued behind a fight holds the level goal in its area (boardHoldsArea)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Malus stand click walks into reach first, burns no clicks out of range (must-fail test)
- [x] #2 Town healing floor capped by the belt plan capacity (Necromancer, Paladin)
- [x] #3 Executor parks a group after MaxRepeats ran-without-progress with a reason
<!-- AC:END -->
