---
id: TASK-513
title: 'verifyp03c2: independent verifier for upstream candidate P03c a knocking hit in a get-hit'
status: In Progress
assignee:
  - verifyp03c2
created_date: '2026-11-15T18:22:00Z'
labels:
  - docs
  - research
dependencies: []
priority: high
ordinal: 278000
card: "verifyp03c2"
owns:
  - "docs/upstream/verifier-p03c-2026-11-15.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P03c (fid/knock-in-gethit-r1) needs an independent Opus pass before posting upstream. Done when: the verifier doc gives a verdict and the exact head to post. Consumer: the P03c posting (docs/upstream/pr-queue.md row 3f).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 rebased r2 on current upstream main pushed to the private fork only
- [ ] #2 claim, oracle, mutant, ctest, clang-tidy, presubmit, pr_guard re-run and recorded
- [ ] #3 verdict and exact head to post in docs/upstream/verifier-p03c-2026-11-15.md
<!-- AC:END -->
