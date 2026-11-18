---
id: TASK-624
title: >-
  decomp39: player melee through walls (FUN_00622c40 reach and line) and the DoT
  sub-point death rule
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T16:20:00Z'
updated_date: '2026-11-17T17:02:00Z'
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
result: "Player melee reach: game.exe starts a swing and lands a blow only within FUN_00622c40 (unit distance within weapon range + extra + 1, no wall of mask 0x804 on the line); d2d used a 1.1 cell (start) and 1.6 cell (blow) disc with no wall test. V run, 101,168 cases x 2: parent 10,009 starts and 33,963 blows differ, claim 0, 8 of 8 mutants abort the pins; fork branch fid/melee-reach-r1 (two commits, pushed to the private remote only), body draft written, verifier pending, whole ctest not run (19 Fight tests pass). Through-wall pairs are 0.1 to 0.6 percent of close pairs outdoors, near zero in caves (static count, U for fights). DoT: game.exe kills under 1 in 256ths (not the hit rule of one point): no deviation of the rule, grain differs under one point (V run, 140 cases). Bot: clicks and lets the game walk; no change. Not verified live; nothing posted."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user saw the d2d Paladin hit monsters through walls; is it game.exe behaviour, how does d2d and the bot differ, and does the DoT death rule of CodeRabbit on #600 exist in game.exe. Done when: the trace doc has the rule with addresses and tags, a proven deviation has a fork branch and a body draft, and the DoT question is answered. Consumer: docs/upstream/pr-queue.md (the verifier queue) and the pal-skill-matrix melee rows.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 trace doc: game.exe rule with addresses and tags, d2d and bot compared
- [x] #2 oracle differential of the player melee start and blow against game.exe (V run)
- [x] #3 DoT sub-point rule answered (V read, V run)
- [x] #4 fork branch, body draft, verifier-ready note
<!-- AC:END -->
