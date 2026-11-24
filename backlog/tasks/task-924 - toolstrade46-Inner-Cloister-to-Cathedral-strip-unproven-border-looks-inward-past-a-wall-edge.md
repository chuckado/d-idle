---
id: TASK-924
title: 'toolstrade46: Inner Cloister to Cathedral strip (unproven border looks inward past a wall edge)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T17:53:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 426000
card: "toolstrade46"
owns:
  - "internal/d2dworld/border.go"
  - "internal/d2dworld/world_test.go"
  - "tools/provingground/scenarios/navexit.json"
exclusive: []
contracts: []
result: "Cause (V live 17:41Z and lab devp27 seed 11): the Cathedral is a border link of the Inner Cloister (34 tiles north, linked only near its edge; the lab warps list of level 32 has one warp, to Jail 3). nav no usable gate: exits 1, adjacent 0, pending 1, unproven 0: the layout edge rows toward the Cathedral are wall, the walkable cells start a few rows inside, so the unproven strip (the cells nav walks to so d2d links the neighbour) was empty. The strip is now taken from the first row within four of the edge that has walkable cells (linked-closed and unlinked neighbours). Lab: nav.inner-cloister-cathedral RED (two runs, same log) then GREEN for Paladin and Necromancer after the fix (devp27, CHEATED setup, job 20261124-125230-9504). Unit test TestUnprovenStripLooksInwardPastAWallEdge fails with the edge row only."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The unproven strip toward a neighbour is read a few rows inside a wall edge, for a linked neighbour with nothing open across and for an unlinked one; must-fail unit test
- [x] #2 Lab scenario nav.inner-cloister-cathedral GREEN for Paladin and Necromancer on devp27 (RED before)
<!-- AC:END -->
