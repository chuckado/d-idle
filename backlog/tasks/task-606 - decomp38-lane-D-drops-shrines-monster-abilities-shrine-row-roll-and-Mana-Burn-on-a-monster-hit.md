---
id: TASK-606
title: >-
  decomp38: lane D drops/shrines/monster abilities: shrine row roll and Mana
  Burn on a monster hit
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T10:03:00Z'
updated_date: '2026-11-17T10:21:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 307000
card: "decomp38"
owns:
  - "docs/research/dc38-lane-d.md"
  - "tools/emu/dc38_shrine.py"
  - "tools/emu/dc38_burn.py"
  - "docs/upstream/bodies/"
exclusive:
  - "fork-branch:fid/mana-burn-amount-r1"
  - "fork-branch:fid/shrine-roll-oracle-r1"
contracts: []
result: "Lane D: shrine row roll FUN_0054f9d0 equals game.exe (200,000 cases, 0 differ, oracle-only branch fid/shrine-roll-oracle-r1). Mana Burn hit: game.exe takes 256 times d2d's mana (stats 62/63 stored x<<8 and rolled with another <<8): 1,750 and 1,800 of 2,000 differ on the parent, 0 after, rule byte-equal 4,000 of 4,000, one mutant caught (fid/mana-burn-amount-r1, body draft written, verifier pending); the monster also heals by the burn (V run game side, no d2d differential). Item drops/chests already traced (treasure-trace.md); other abilities not run. Offline: emulator and fork builds on the builder; no live game; nothing posted."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: lane D of the decomposition wave (drops, shrines and objects, monster abilities); pick the top three by impact on the bot times tractability and prove or clear each against game.exe. Done when: the trace doc names the rules with addresses and tags, a proven deviation has a private fork branch and a body draft, and the rest is a negative result. Consumer: docs/upstream/pipeline-2026-11-14.md rows P14 (shrines) and the fidelity matrix monster rows.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 traced topics with FUN_ addresses and V run oracles
- [x] #2 proven deviation has a private fork branch and a body draft, or a negative result
<!-- AC:END -->
