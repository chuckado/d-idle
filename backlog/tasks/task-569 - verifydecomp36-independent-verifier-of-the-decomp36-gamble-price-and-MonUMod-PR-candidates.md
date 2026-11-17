---
id: TASK-569
title: >-
  verifydecomp36: independent verifier of the decomp36 gamble price and MonUMod
  PR candidates
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-16T22:52:00Z'
updated_date: '2026-11-17T01:56:00Z'
labels:
  - d2d
  - readonly
dependencies:
  - TASK-558
priority: high
ordinal: 297000
card: "verifydecomp36"
owns:
  - "docs/upstream/verifier-decomp36-2026-11-16.md"
exclusive: []
contracts: []
result: "gamble r1 FAIL, r2 PASS WITH CHANGES (fid/gamble-price-r2 f9a127e1+48d8f6ee on upstream def80df7: 48,310 then 0 of 73,062 differ, reduced 1..99 grid 0); MonUMod r1 FAIL, r3 PASS WITH CHANGES (fid/umod-can-take-r3 7bd3ddec+fdea73d1: 11 then 0 of 31,562; r2 had a tidy finding); ctest same 2 failures as upstream, mutants abort the r2 tests; the 11 monsters are uniques in Acts 1-3 and 5, not Acts 3-5 only; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Opus verifier pass on fid/gamble-price-r1 and fid/umod-can-take-r1 (decomp36, TASK-558). Nothing posted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Game rules re-derived (gamble normcode, MonUMod nomultishot), callers and other price paths checked
- [x] #2 Clean-identity oracle-first r2 branches on the private fork remote
- [x] #3 Builds, ctest against upstream, differentials, mutants, usage errors, tidy, pr_guard, presubmit run on the NAS builder
- [x] #4 Verdict per PR, required body changes, exact branches and commits to post in the verifier doc
<!-- AC:END -->
