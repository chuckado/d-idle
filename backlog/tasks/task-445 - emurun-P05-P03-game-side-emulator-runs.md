---
id: TASK-445
title: 'emurun: P05/P03 game-side emulator runs'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T21:36:00Z'
updated_date: '2026-11-14T21:37:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 259000
card: "emurun"
owns:
  - "docs/research/emu-runs-2026-11-14.md"
  - "tools/emu/fl_gold.py"
  - "tools/emu/block_len.py"
  - "tools/emu/emurun_gold.sh"
  - "tools/emu/emurun_block.sh"
exclusive: []
contracts: []
result: "P05 game column = rule on 539 of 539 (V run); P03 game vs d2d 17,195 of 34,314 differ, game = closed form on 34,314 (V run). After-side and P02 not run. Not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fl_gold.py game run vs rule
- [x] #2 block_len.py game run vs d2d and closed form
- [x] #3 record written
<!-- AC:END -->
