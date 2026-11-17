---
id: TASK-624
title: 'decomp39: player melee through walls (FUN_00622c40 reach and line) and the DoT sub-point death rule'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T16:20:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 311000
card: "decomp39"
owns:
  - "docs/research/decomp39-melee-walls.md"
  - "tools/emu/dc39_reach.py"
  - "tools/emu/dc39_dot.py"
  - "docs/upstream/bodies/decomp39-melee-reach-body.md"
exclusive:
  - "fork-branch:fid/melee-reach-r1"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 trace doc: game.exe rule with addresses and tags, d2d and bot compared
- [ ] #2 oracle differential of the player melee start and blow against game.exe (V run)
- [ ] #3 DoT sub-point rule answered (V read, V run)
- [ ] #4 fork branch, body draft, verifier-ready note
<!-- AC:END -->
