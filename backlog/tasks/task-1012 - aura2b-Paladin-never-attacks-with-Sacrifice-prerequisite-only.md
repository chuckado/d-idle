---
id: TASK-1012
title: 'aura2b: Paladin never attacks with Sacrifice (prerequisite only)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T18:29:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 485000
card: "aura2b"
lane: next
owns:
  - "internal/combat/classdata/profiles.json"
  - "internal/combat/rules.go"
  - "internal/combat/class_test.go"
  - "internal/combat/lifeprice_test.go"
  - "internal/agency/derive.go"
  - "internal/agency/testdata/agreement.json"
exclusive: []
result: "Unit-tested only, not live. Default and hammerdin Paladin profiles: Sacrifice dropped from the bands and from the life-price list, listed in avoid; early bands now Attack, Zeal from clvl 12 once learned. Smite and Holy Bolt (undead only) stay unranked per the SME note. The planner agreement fixture moved to 44 of 60."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: SME (docs/sme/class-paladin.md, paladin-auras 2026-11-26): Sacrifice costs life per swing, 1 point is a Zeal prerequisite only. The default Paladin profile ranked it as the clvl 2 to 14 attack (Questy ran it at clvl 12). Done when: no band ranks it, no life price, Zeal from clvl 12 else the weapon, the planner skips it.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Neither Paladin profile ranks Sacrifice or pays life; the planner skips avoided attacks
- [x] #2 Tests for two Paladin profiles over clvl 1 to 40
<!-- AC:END -->
