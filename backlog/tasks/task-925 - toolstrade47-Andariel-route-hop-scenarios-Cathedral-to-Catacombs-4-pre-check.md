---
id: TASK-925
title: >-
  toolstrade47: Andariel route hop scenarios (Cathedral to Catacombs 4),
  pre-check
status: Done
assignee:
  - credit-runner
created_date: '2026-11-24T18:12:00Z'
updated_date: '2026-11-24T18:23:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 427000
card: "toolstrade47"
owns:
  - "tools/provingground/scenarios/navexit.json"
exclusive: []
contracts: []
result: "V lab devp27 seed 11 (CHEATED setup, jobs 20261124-130035-9960 and -130602-31d7): every hop of the Andariel route is a plain stair warp with a click point (exitsread.py: 33 to 34 at tile (21,10), 34 to 35 at (54,3), 35 to 36 at (66,3), 36 to 37 at (6,3); no 0x0 select box; Catacombs 4 has only the way back, Andariel is in the same level). Real ExitTo: 33 to 34, 34 to 35, 36 to 37 GREEN for Paladin and Necromancer; 35 to 36 GREEN for Paladin (15 doors in a row, the old 8 attempt limit would have failed it) and RED for the Necromancer, who died in the Catacombs 2 crowd (21 hostiles near the first door, three runs): a combat finding, no exit failure. Door counts per hop: up to 15 (35 to 36), 13 (34 to 35), 7 (36 to 37), all under the cap of 24."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Lab scenarios for the four stair hops 33 to 37 and their verdicts
<!-- AC:END -->
