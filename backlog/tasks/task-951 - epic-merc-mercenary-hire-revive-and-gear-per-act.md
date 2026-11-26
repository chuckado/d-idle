---
id: TASK-951
title: 'epic-merc: mercenary hire, revive and gear per act'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-26T03:05:00Z'
labels:
  - bot
  - feature
  - T4
dependencies: []
priority: medium
ordinal: 452000
card: "epic-merc"
owns:
  - "docs/epics/epic-merc.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Real gap evidenced in the repo: the lab merc is a level 1 Rogue that dies at once (matrix note 1; `debug merc hire` has no level argument, section 8); `drill.tools.merc17` RED; Andariel readiness reads `dead merc` (matrix note 69, merc died 3 times in the bold kill). Policy exists (docs/merc.md, merc-revive.md, merc-wire.md, merclive.md) behind merc.enabled. Done when: the criteria below are ticked.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Hire per act: the Act 1 Rogue, Act 2 desert mercenary, Act 3 and Act 5 hire rows work through the town planner, verified live (docs/merc.md, docs/merclive.md).
- [ ] #2 Revive: a dead merc is revived in town when gold allows, drilled (docs/merc-revive.md, matrix note 69 `dead merc`).
- [ ] #3 Gear: the merc is geared from found or bought items per act, and the gear survives replace (docs/merc.md gear policy).
- [ ] #4 Lab verb: `debug merc hire` takes a level or the drills use a level-appropriate merc, so Andariel and Tools rows are not judged with a level 1 merc (matrix section 8).
- [ ] #5 A fresh-hero live run shows hire, revive and gear working with writes after start 0.
<!-- AC:END -->
