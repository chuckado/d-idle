---
id: TASK-809
title: >-
  mapverify: Countess three ways with the map knobs, the Monastery Gate border
  strip, default-flip memo
status: Done
assignee:
  - credit-runner
created_date: '2026-11-22T16:29:00Z'
updated_date: '2026-11-22T18:58:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 369000
card: "mapverify"
owns:
  - "docs/probes/countess-knobs-2026-11-22.md"
  - "docs/probes/strip-monastery-2026-11-22.md"
  - "tools/probes/strip.py"
  - "tools/probes/knobsummary.py"
  - "tools/provingground/knobs3way.sh"
  - "cmd/d2scenario/run.go"
  - "cmd/d2scenario/envknobs_test.go"
  - "internal/d2dworld/border.go"
  - "internal/d2dworld/world.go"
  - "internal/d2dworld/placelinked_test.go"
  - "internal/nav/gatedoor.go"
  - "internal/nav/gatedoor_test.go"
  - "internal/nav/nav.go"
  - "internal/nav/exit.go"
  - "tools/probes/objclass/main.go"
exclusive: []
contracts: []
result: "Paladin 11/12/13 GREEN three ways on lab slots (CHEATED setups): nav.fullGrid identical to off to the click, nav.warpRead mouth-gate refusals 55 to 0 and 28 percent fewer clicks and goal time; Sorceress BLOCKED (out of mana), so no caster data. Monastery Gate to Outer Cloister: the layout verb never lists the cloister chain (placeLinked from debug nearby) and two closed Door objects shut the gate (nav opens them, World.Refetch, invalidate fixed): act1.andariel.slaughter Paladin 11 RED to GREEN on the lab d2d, Sorceress 3 RED (never reaches the gate). Verified on lab slots, not on a real game or the sandbox. Defaults not flipped; memo in countess-knobs doc."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Done when: the three-way table and the memo exist and the strip cause is fixed or named.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 act1.forgotten.countess run three ways (knobs off, nav.fullGrid, fullGrid + warpRead) on lab slots, table in docs/probes/countess-knobs-2026-11-22.md
- [x] #2 Monastery Gate to Outer Cloister border strip probed, cause recorded, fix with a must-fail test
- [x] #3 default-flip memo for nav.fullGrid and nav.warpRead, and what is missing for quests.boardSM
- [x] #4 go vet, gofmt, race -count=20 on d2dworld and d2scenario, check.sh, leakcheck clean
<!-- AC:END -->
