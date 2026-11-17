---
id: TASK-607
title: >-
  decomp37: combat and skills lane C, a player's gear absorb, Open Wounds and
  the magic resist (elemental pipeline)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T10:09:00Z'
updated_date: '2026-11-17T12:27:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 307000
card: "decomp37"
owns:
  - "docs/research/decomp37-combat.md"
  - "tools/emu/dc37_"
  - "docs/upstream/bodies/decomp37-"
exclusive: []
contracts: []
result: "Elemental pipeline lane C, three topics against game.exe 1.14d: (T1) a player's gear absorb (stats 142..149, FUN_0057bf10 after the resistance: percent capped at 40, then flat): d2d read none; fid/gear-absorb-r2 (oracle 00e7b018, claim 82eced5f) 13,189 of 20,000 and 39,678 of 60,000 differ on the parent, 0 on the claim, 5 mutants caught by the compare and by test_combat; (T2) Open Wounds handler found (FUN_005bfe60, event func 15): d2d's level bracket equals game.exe, but it never halves on a champion/unique/superunique and cuts to whole points a second: fid/open-wounds-r2 (oracle 857ece45, claim 6d4b506d) 1,748 of 2,000 and 16,463 of 21,782 differ, 0 after, 4 mutants abort test_combat; (T3) the player's magic resist on a magic blow: 1,550 of 3,000 differ (V run), not built. test_combat ok and test_game rc 0 on both claims (builder); full ctest, presubmit, clang-tidy not run (the verifier does). Bot finding B1: resist capped before the difficulty penalty (internal/combat/rules.go:311, internal/readiness/model.go:357). Branches pushed to the private fork remote only; nothing posted. Verified in the emulator and on builds, not live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Consumer: docs/upstream/pr-queue.md (the verifier queue of the upstream PR pipeline); lane C of the 2026-11-17 decomposition wave. Docs and tools/emu only; nothing posted upstream.
Size: L

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Three untraced combat topics traced against game.exe with V run tags
- [x] #2 Each proven deviation has a private fork branch and a PR body draft, or a negative result with evidence
- [x] #3 check.py, contracts, leakcheck clean
<!-- AC:END -->
