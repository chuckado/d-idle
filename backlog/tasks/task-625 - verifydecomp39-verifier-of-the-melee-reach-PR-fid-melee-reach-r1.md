---
id: TASK-625
title: 'verifydecomp39: verifier of the melee reach PR (fid/melee-reach-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T17:01:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies:
  - TASK-624
priority: high
ordinal: 312000
card: "verifydecomp39"
owns:
  - "docs/upstream/verifier-decomp39-2026-11-17.md"
exclusive: []
contracts: []
result: "r1 FAIL (pr_guard label, short names, include-cleaner, usage errors, pin gap), r2 PASS WITH CHANGES: fid/melee-reach-r2 60672f4e+39f9cda6 on upstream 895f8a67, pushed to the private fork; parent 2,799/9,869 (seed 1), 7,210/24,094 (seed 7), 2,835/9,868 (fresh seed 200) start/blow differ, claim 0 on all; ctest 359/359; 11 mutants (m4 equivalent, tentacle m9 survives, rest abort the 18 pins); stall harness: 28 of 400 close wall-cut Blood Moor pairs never land a blow after the claim (game.exe U); monsters, merc, pets keep wall-blind reach; body changes B1-B11; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Verdict PASS / PASS WITH CHANGES / FAIL with the branch and commits to post
- [x] #2 Full ctest on the claim (none fails, so no upstream run), mutants, reseed, tidy, pr_guard, presubmit
- [x] #3 Fight-stall, monster, ranged, merc scope questions answered with V or U tags
<!-- AC:END -->
