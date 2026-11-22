---
id: TASK-809
title: 'mapverify: Countess three ways with the map knobs, the Monastery Gate border strip, default-flip memo'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T16:29:00Z'
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
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Done when: the three-way table and the memo exist and the strip cause is fixed or named.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 act1.forgotten.countess run three ways (knobs off, nav.fullGrid, fullGrid + warpRead) on lab slots, table in docs/probes/countess-knobs-2026-11-22.md
- [ ] #2 Monastery Gate to Outer Cloister border strip probed, cause recorded, fix with a must-fail test
- [ ] #3 default-flip memo for nav.fullGrid and nav.warpRead, and what is missing for quests.boardSM
- [ ] #4 go vet, gofmt, race -count=20 on d2dworld and d2scenario, check.sh, leakcheck clean
<!-- AC:END -->
