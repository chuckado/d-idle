---
id: TASK-1011
title: >-
  aura2a: read the real active aura on d2d, re-pick a stale one, aura-off
  invariant
status: Done
assignee:
  - credit-runner
created_date: '2026-11-26T18:22:00Z'
updated_date: '2026-11-26T18:56:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 485000
card: "aura2a"
lane: next
owns:
  - "internal/combat/aura.go"
  - "internal/combat/class_test.go"
  - "internal/combat/fighter.go"
  - "internal/d2/player.go"
  - "internal/d2dsrc/source.go"
  - "internal/d2dsrc/source_test.go"
  - "internal/bot/invariants.go"
  - "internal/bot/invariants_test.go"
  - "internal/bot/observe.go"
  - "cmd/d2scenario/invariants.json"
  - "docs/invariants.md"
exclusive: []
result: "Unit-tested only (combat, bot, d2dsrc, d2, d2scenario on the worker host); not verified live. The d2d state line aura= feeds ActiveAura, a stale aura is re-picked through another skill, aura-off is a soft row, the minute line carries aura."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the bot trusted the selected right skill as the aura state; d2d's state line aura= is the exact state (V live, auralive probe). Done when: KeepAura compares the real aura, forces a re-pick via another skill when stale, logs only confirmed ups, and the aura-off row watches it. Unit-only: not run on a lab slot; the load case is U (checklist in the report).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 ActiveAura on d2d reads aura= from the state line; KeepAura re-picks a stale aura
- [x] #2 aura up logged only when the state confirms; aura missing warning once a minute; minute line aura key
- [x] #3 aura-off soft invariant with must-fail tests for two heroes
<!-- AC:END -->
