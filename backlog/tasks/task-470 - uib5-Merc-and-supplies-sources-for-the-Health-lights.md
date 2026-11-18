---
id: TASK-470
title: 'uib5: Merc and supplies sources for the Health lights'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T04:51:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 266000
card: "uib5"
owns:
  - "cmd/d2bot/glancewire.go"
  - "cmd/d2bot/glancewire_test.go"
  - "cmd/d2bot/live.go"
  - "cmd/d2bot/d2dbackend.go"
exclusive: []
contracts: []
result: "Merc and Supplies lights bound by useGlance; tested for Sorceress and Paladin; not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: /api/glance reports the Mercenary and Supplies lights as unknown until the bot binds sources. Done when: Hub.SetGlanceSources is called by d2bot (useGlance, two one-line call sites in live.go and d2dbackend.go) with Merc (the merc desk state alive/dead/none) and Supplies (belt healing/mana against the belt plan columns x belt rows of resource.PlanForBuild; portal source) readers; glanceExtra lets B6 add the Needs source without a second SetGlanceSources call.

Live check (not run): in town with a full belt the Supplies light is ok; drink to below plan: warn; empty belt or no portal: bad; hire/kill the merc and watch the Merc light alive/dead/none on GET /api/glance.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Tests for two classes with fake snapshots: each light value; nil snapshot -> unknown
- [x] #2 No existing glance golden changes (the lights are asserted in cmd/d2bot tests; no watch golden touched)
<!-- AC:END -->
