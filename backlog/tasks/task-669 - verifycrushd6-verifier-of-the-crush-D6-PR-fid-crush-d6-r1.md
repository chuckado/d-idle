---
id: TASK-669
title: 'verifycrushd6: verifier of the crush D6 PR (fid/crush-d6-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T17:16:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies: []
priority: high
ordinal: 328000
card: "verifycrushd6"
owns:
  - "docs/upstream/verifier-crushd6-2026-11-19.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: independent Opus check of fid/crush-d6-r1 (decomp42) before posting upstream. Consumer: the posting agent (docs/upstream/pr-queue.md). Done when: the verdict doc names the branch and commits to post and the body changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Rule re-derived from the decompilation for every path through FUN_0057c6c0; d2d's other crush and get-hit users checked
- [ ] #2 Rebased on current upstream main and on D5; build, full ctest vs parent, oracle, mutants incl. missile and element, tidy, pr_guard, presubmit on the builder
- [ ] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
