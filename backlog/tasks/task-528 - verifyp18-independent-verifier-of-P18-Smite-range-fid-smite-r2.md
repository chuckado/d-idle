---
id: TASK-528
title: 'verifyp18: independent verifier of P18 Smite range (fid/smite-r2)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T02:30:00Z'
updated_date: '2026-11-16T05:45:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: high
ordinal: 283000
card: "verifyp18"
owns:
  - "docs/upstream/verifier-p18-2026-11-15.md"
exclusive: []
contracts: []
result: "PASS WITH CHANGES (offline, NAS builder; not live). r2 (rebased on upstream 485fe535) reproduces: oracle 63,709 of 75,600 differ, claim 0, mutants (no width / no stat 111) abort; ctest 5 known failures on both. But r2 test_combat is seed-dependent (116/200 reseeds fail, upstream 0/200): post fid/smite-r3 (oracle bb5cc4d0, claim 8480acb8; 0/200) after the body fixes (Checklist + skills.md line, stats 17/18 limit, reach wording). Doc docs/upstream/verifier-p18-2026-11-15.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Consumer: upstream-pr P18 (smite-range). Opus verifier before posting P18 (Smite: stat 111 and max >= min + 1 point). Done when the verifier doc gives a verdict and the head to post.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fid/smite-r2 rebased on upstream main, rebuilt, grid compare on oracle and claim, mutants, full ctest vs upstream main
- [x] #2 Smite range rule re-derived from game.exe (FUN_0057b420, FUN_005ce9f0), grid axes and test_combat bounds judged
- [x] #3 Body honesty checked; verdict and exact head in docs/upstream/verifier-p18-2026-11-15.md
<!-- AC:END -->
