---
id: TASK-562
title: 'survfix2: realized yield, mana floor, fresh guard, merc revive loop, Paladin weapon in the A/B hero'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-16T19:38:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 294000
card: "survfix2"
owns:
  - "internal/survival"
  - "internal/combat/castfirst.go"
  - "internal/combat/fight.go"
  - "internal/combat/fightend.go"
  - "internal/combat/dodge.go"
  - "internal/bot/zonemodel.go"
  - "internal/bot/zoneshadow.go"
  - "internal/bot/zonemodel_test.go"
  - "internal/env/tuning_bot.go"
  - "cmd/d2bot/survivalwire.go"
  - "tools/ab/prep"
  - "tools/survival/abrun.sh"
  - "docs/survival/ab-harness.md"
  - "docs/survival/ledger.md"
  - "docs/knobs.md"
exclusive: []
contracts: []
result: "Offline, not verified live. With bot.zoneModel 1 only: realized xp/min blends into the zone value (trust floor, per-archetype data), mana-starved zones lose value, fresh_game refused with unread offence or dead merc, merc revive loop breaker (bot.mercReviveLoopN 3, bot.mercReviveLoopMin 10); revive decisions and time-to-death always logged. A/B hero gets a SETUP weapon (hax) so the Paladin reads src=skill (slot 4 U until a run). Tests on Paladin and Sorceress. zoneModel default stays 0."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Implements the proposals of docs/survival/ab-screen-analysis-2026-11-16.md section 6. Default bot.zoneModel stays 0; new behaviour only with zoneModel 1. Done when the survival, bot and tools tests pass.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Realized xp/min blends into the zone value (data-driven, two builds tested)
- [x] #2 Observed mana floor lowers a starving zone
- [x] #3 fresh_game refused with unread offence or dead merc
- [x] #4 Merc revive logging and revive-loop breaker (bot.mercReviveLoopN, bot.mercReviveLoopMin)
- [x] #5 A/B harness hero gets a weapon (declared SETUP), docs updated
<!-- AC:END -->
