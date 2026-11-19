---
id: TASK-558
title: 'decomp36: lane B items/town/monster: death penalty, gamble price and stock, unique mod eligibility'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T18:42:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 292000
card: "decomp36"
owns:
  - "docs/research/dc36-lane-b.md"
  - "tools/emu/dc36_death.py"
  - "tools/emu/dc36_gamble.py"
  - "tools/emu/dc36_umod.py"
  - "tools/emu/dc36_cmp.py"
  - "docs/upstream/bodies/"
exclusive:
  - "fork-branch:fid/death-corpse-exp-r1"
  - "fork-branch:fid/gamble-price-r1"
  - "fork-branch:fid/umod-can-take-r1"
contracts: []
result: "Lane B: death penalty (240,000 cases, game.exe vs d2d: gold, stash, pile, goldlost, exp equal; corpse-told exp 1 low in the clamp, trivial, fid/death-corpse-exp-r1), gamble price (48,310 of 73,062 differ on an upgraded base, 0 after, fid/gamble-price-r1, economy effect), MonUMod eligibility (11 of 31,562 pairs, Multiple Shots on nomultishot rows, 0 after, fid/umod-can-take-r1). Offline, V run on the builder emulator and fork builds; no live game; nothing posted; verifier not yet run."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: lane B of the decomposition wave (items and economy, town and world, monster AI): trace three topics nobody had run against game.exe and prove or clear each deviation with the existing emulator differential. Done when: the trace doc names the rule with addresses and tags, each proven deviation has a private fork branch and a body draft, and the rest is a negative result. Consumer: docs/upstream/pipeline-2026-11-14.md rows P20..P22 and the fidelity matrix cells PR08, VN04, MA06.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 three un-traced topics traced with FUN_ addresses and V run oracles
- [x] #2 each proven deviation has a new private fork branch and a body draft; else a negative result
<!-- AC:END -->
