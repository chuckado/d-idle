---
id: TASK-919
title: >-
  toolstrade42: hold_timeout keeps the readiness reason, Andariel gate carries
  the Barracks floor
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T13:34:00Z'
updated_date: '2026-11-24T14:02:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 423000
card: "toolstrade42"
owns:
  - "internal/questboard/sm.go"
  - "internal/questboard/exec_sm.go"
  - "internal/questboard/exec_sm_test.go"
  - "internal/task/gate.go"
  - "internal/task/route_test.go"
exclusive: []
contracts: []
result: "Cause of the live text: the stored hold_timeout park was rendered from the bare code with the run counters (n = failures = 0) and an unformatted {hold}; the run now stores the hold code, its words and the budget, so it reads: Still waiting after 60 minutes: Too hard for now: level 24 needed for the Barracks (clvl 20). Training toward it. Andariel (floors from data: Paladin 17, merc 15, rec 20 and 18; Necromancer 17, 15, 18, 16; route 14; kit TP, potions, 3 mana for a caster, 2 antidotes) passed its gate at clvl 20 while the road to her (Barracks) is closed below 24 melee, 22 caster and max life 400 or 300: the gate now carries the Barracks floor, so the board shows level N needed instead of starting. Offline tests only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 hold_timeout park reads the hold's readiness reason and the budget, in memory and over a restart (two classes)
- [x] #2 Andariel gate carries the Barracks floor so the board says level N needed (Paladin, Necromancer)
<!-- AC:END -->
