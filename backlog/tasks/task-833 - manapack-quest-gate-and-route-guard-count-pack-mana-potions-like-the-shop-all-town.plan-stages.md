---
id: TASK-833
title: 'manapack: quest gate and route guard count pack mana potions like the shop (all town.plan stages)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T01:44:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 380000
card: "manapack"
owns:
  - "internal/bot/routeguard.go"
  - "internal/bot/kitplan_test.go"
exclusive: []
contracts: []
result: "fillKit now counts mana in belt plus pack at every stage (was belt only at stage 0 and 1; the sandbox runs quests.boardSM=1, which already took the pack path, so the friend board is not affected today; any run with boardSM 0 and town.plan below 2 was). Unit tests fail without the fix. Scenario rerun: see the card note."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 World.ManaPotions counts belt plus pack at stage 0 and 1 too
- [x] #2 Must-fail tests on Necromancer, Sorceress, Paladin with potions only in the pack
<!-- AC:END -->
