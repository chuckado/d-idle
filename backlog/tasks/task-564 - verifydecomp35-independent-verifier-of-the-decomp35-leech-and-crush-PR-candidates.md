---
id: TASK-564
title: >-
  verifydecomp35: independent verifier of the decomp35 leech and crush PR
  candidates
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T19:40:00Z'
updated_date: '2026-11-16T22:49:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 294000
card: "verifydecomp35"
owns:
  - "docs/upstream/verifier-decomp35-2026-11-16.md"
exclusive: []
contracts: []
result: "leech r2 PASS WITH CHANGES (fid/leech-r2 26ae7020+55cb4e68 on upstream c3dc51c3); crush r1 FAIL (leech cut before the crush: 1,068/2,810 and 6,589/20,810 differ), r2 PASS WITH CHANGES (fid/crush-r2 1a75fed3+e5dc17d3, post after leech merges); ctest same 2 of 342 as upstream, mutants abort test_combat; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Verdict per PR with required body changes and the exact branches and commits to post
- [x] #2 Builds, differentials, mutants and ctest re-run on the builder
<!-- AC:END -->
