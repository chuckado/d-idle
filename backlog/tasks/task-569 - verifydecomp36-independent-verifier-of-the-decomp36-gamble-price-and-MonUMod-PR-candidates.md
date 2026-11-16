---
id: TASK-569
title: >-
  verifydecomp36: independent verifier of the decomp36 gamble price and MonUMod
  PR candidates
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-16T22:52:00Z'
updated_date: '2026-11-16T22:52:00Z'
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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Opus verifier pass on fid/gamble-price-r1 and fid/umod-can-take-r1 (decomp36, TASK-558). Nothing posted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Game rules re-derived (gamble normcode, MonUMod nomultishot), callers and other price paths checked
- [ ] #2 Clean-identity oracle-first r2 branches on the private fork remote
- [ ] #3 Builds, ctest against upstream, differentials, mutants, usage errors, tidy, pr_guard, presubmit run on the NAS builder
- [ ] #4 Verdict per PR, required body changes, exact branches and commits to post in the verifier doc
<!-- AC:END -->
