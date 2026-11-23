---
id: TASK-849
title: 'toolstrade5: mid-route kit floors, border strip probe'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T04:02:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 384000
card: "toolstrade5"
owns:
  - "internal/task/route.go"
  - "internal/task/route_test.go"
  - "internal/bot/routeguard.go"
  - "tools/provingground/scenarios/act1more.json"
  - "tools/probes/strips.py"
  - "tools/probes/strips_test.py"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "guardedExit uses a mid-route verdict (halved potion floors, same level rule); strips.py probe written, run on the lab not yet; chunk green not shown."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. The Outer Cloister arrival: the route guard sent the hero home at 3 healing potions of 4, the portal was refused by the crowd rule and the chunk parked. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A hero on the guarded road tolerates a thinned kit (halved potion floors) before it turns round; must-fail test, two classes
- [x] #2 tools/probes/strips.py lists the approach cell of every border link (unit-tested flood; lab run pending)
<!-- AC:END -->
