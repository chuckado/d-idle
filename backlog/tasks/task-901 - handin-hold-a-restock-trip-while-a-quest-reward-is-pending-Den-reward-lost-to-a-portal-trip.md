---
id: TASK-901
title: >-
  handin: hold a restock trip while a quest reward is pending (Den reward lost
  to a portal trip)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T21:07:00Z'
updated_date: '2026-11-26T01:45:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 407000
card: "handin"
owns:
  - "internal/bot/goal.go"
  - "internal/bot/runner.go"
  - "internal/bot/questboard_test.go"
  - "tools/provingground/scenarios/necro.json"
  - "docs/probes/necromancer-act1-drill-2026-11-22.md"
exclusive: []
contracts: []
result: "holdTripForReward: a non-emergency restock trip waits (max 2 virtual minutes) while the live quest reader lists a quest as cleared with its reward pending; unit test; necro.act1-stage1-nomp GREEN on the lab (job 20261123-155622-a226), the same fixture was RED before (7536). Baseline on identical main not run (builder unreachable): U for other merges. Stage 2b still RED (Smith killed by the army, Malus not picked up, kit gate)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A pending quest reward holds a non-emergency restock trip for at most 2 minutes, unit-tested
- [x] #2 Stage 1 without a mana potion GREEN on the lab
<!-- AC:END -->
