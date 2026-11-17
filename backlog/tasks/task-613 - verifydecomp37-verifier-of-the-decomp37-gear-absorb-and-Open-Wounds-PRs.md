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
result: ""
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
- [ ] #1 Verdict per PR with required body changes and the exact branch and commits to post
- [ ] #2 Rules re-derived from the decomp, mutants, reseed, ctest, tidy, presubmit run and recorded
<!-- AC:END -->
