---
id: TASK-613
title: 'verifydecomp37: verifier of the decomp37 gear absorb and Open Wounds PRs'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-17T12:58:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 309000
result: "Verifier of decomp37 (docs/upstream/verifier-decomp37-2026-11-17.md). Gear absorb fid/gear-absorb-r2: FAIL, the reduction rule is exact (13,189 and 13,242 of 20,000 differ on the parent, 0 after) but game.exe also heals the defender by the absorbed amount (FUN_0057a980) and the merc is changed too; needs a worker r3 (A1..A6), nothing to post. Open Wounds r2 FAIL (tests and oracle miss the path), fid/open-wounds-r3 (oracle 5e7a69b1, claim dabc3ff4, pushed to the private fork) PASS WITH CHANGES W1..W7: new test_open_wounds through Fight::land and monster_dots fails on the parent, every mutant tried aborts, ctest 356/356 (two UI tests flaky under load), tidy 0. Builder and emulator runs, not live; nothing posted."
card: "verifydecomp37"
owns:
  - "docs/upstream/verifier-decomp37-2026-11-17.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the two decomp37 PR candidates (fid/gear-absorb-r2, fid/open-wounds-r2; TASK-607) need an independent verifier before posting. Done when: docs/upstream/verifier-decomp37-2026-11-17.md gives a verdict per PR, the required body changes and the branch and commits to post. Consumer: TASK-607 (the decomp37 PR bodies) and the upstream PR queue.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Verdict per PR with required body changes and the exact branch and commits to post
- [x] #2 Rules re-derived from the decomp, mutants, reseed, ctest, tidy, presubmit run and recorded
<!-- AC:END -->
