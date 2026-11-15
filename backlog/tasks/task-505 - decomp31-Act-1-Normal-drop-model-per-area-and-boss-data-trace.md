---
id: TASK-505
title: 'decomp31: Act 1 Normal drop model per area and boss (data + trace)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T15:32:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 277000
card: "decomp31"
owns:
  - "docs/research/decomp31-drops.md"
  - "internal/plandata/dropmodel-normal.json"
  - "tools/emu/dm_"
exclusive: []
contracts: []
result: "V run (game.exe roll under the emulator, 117 keys x 20,000 rolls, 12 population seeds, 36 Act 1 levels); d2d equal on 40,800 rolls; no PR candidate; nothing live; no Go changed"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. Drop model as data for levelsim and loot rules. Consumer: levelsim -drops, the Where-to line and the loot-rules examples (follow-up cards in docs/research/decomp31-drops.md section 9). Done when the doc and the JSON are merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Per-area and boss drop model computed from game.exe code under the emulator
- [x] #2 docs/research/decomp31-drops.md and internal/plandata/dropmodel-normal.json written
- [x] #3 d2d deviations compared, follow-up cards listed
<!-- AC:END -->
