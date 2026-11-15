---
id: TASK-499
title: 'decomp28: portal-cast interrupt rule, bot findings B2-B4, danger units fit'
status: Proposed
assignee:
  - pro-runner
created_date: '2026-11-15T14:55:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 276000
card: "decomp28"
owns:
  - "docs/research/decomp28-trace.md"
  - "docs/research/data28"
  - "docs/survival/units-fit-2026-11-15.md"
  - "internal/danger/areaunits.json"
  - "tools/emu/dc28_"
exclusive: []
contracts: []
result: "docs/research/decomp28-trace.md: no hit, block or knock-back interrupts a TP or Identify cast on game.exe (V run, 14,400 cases): the bot's R6 premise is false, the guard is over-cautious but right in effect (design card, nothing changed); B4 is a real defect (block credited against unblockable ranged fire in internal/power/margin.go:181); danger units fit from game.exe rooms: shipped 94 percent median error over 67 Act 1 and 2 areas, tiles model 17 percent leave-one-out (presets 54 not fixable), PerScreen 10 to 30 times too low in mazes. Offline only, verified on the emulator, not live; no Go change."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Consumer: survival reflex layer (R6) and the danger generator

Why: the bot's crowd guard rests on a cut-cast premise and the danger generator misses its Act 1 fit (41 percent) and is far worse on Act 2. Done when: the trace, the fit report and the data file are on main.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Town Portal / long cast interrupt rule traced in game.exe with an oracle run (V run)
- [x] #2 Bot crowd guard (R6) judged against the rule; design card proposed, no Go change
- [x] #3 B2 to B4 of decomp26 listed and verified
- [x] #4 Danger units fit: game.exe room population for Act 1 and Act 2 Normal, held-out fit, data file
<!-- AC:END -->
