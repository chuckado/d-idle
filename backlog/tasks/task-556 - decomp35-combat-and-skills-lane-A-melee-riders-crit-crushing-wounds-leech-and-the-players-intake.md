---
id: TASK-556
title: 'decomp35: combat and skills lane A, melee riders (crit, crushing, wounds, leech) and the player''s intake'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T18:35:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 292000
card: "decomp35"
owns:
  - "docs/research/decomp35-combat.md"
  - "tools/emu/d35_"
  - "docs/upstream/bodies/decomp35-"
exclusive: []
result: "Melee riders traced with V run on game.exe: (T1) a player's leech: 256ths, cut to the target's life, divided by DifficultyLevels LifeStealDivisor (1/2/3): fid/leech-r1 (head 99679ead) 17,097 of 20,351 and 50,626 of 60,351 differ on the parent, 0 on the claim, 4 mutants caught; (T2) Crushing Blow takes the life before the hit (events 5/1 run before LAB_0057c852, 2,360 of 2,360), fid/crush-r1 (head de243e2b, stacked) 2,223 of 2,360 differ, 0 after, 3 mutants caught; (T3) gear element rolls: proven but small (2,223 of 3,947 differ in whole points, mean +1 percent), no branch; open wounds U. Both branches ready for the verifier (not run). Offline: emulator and builds on the NAS builder; no game; nothing posted."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

User 2026-11-16 lane A. Docs and tools/emu only; nothing posted upstream. Consumer: docs/upstream/pr-queue.md (the verifier queue of the upstream PR pipeline).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Three un-traced combat topics traced against game.exe with V run tags
- [x] #2 Each proven deviation has a private fork branch and a PR body draft, or a negative result with evidence
- [x] #3 check.py, contracts, leakcheck clean
<!-- AC:END -->
