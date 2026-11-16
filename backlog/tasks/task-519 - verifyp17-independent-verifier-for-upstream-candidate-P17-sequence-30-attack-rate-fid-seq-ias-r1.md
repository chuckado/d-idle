---
id: TASK-519
title: 'verifyp17: independent verifier for upstream candidate P17 (sequence -30 attack rate, fid/seq-ias-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T01:31:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 282000
card: "verifyp17"
owns:
  - "docs/upstream/verifier-p17-2026-11-15.md"
exclusive: []
contracts: []
result: "PASS WITH CHANGES (not verified live; offline on the NAS builder). r2 fid/seq-ias-r2 (oracle 2b22bee5, claim f6fc6a52) rebased on upstream 485fe535 reproduces every count: parent 6,941 of 7,600 differ, claim 0; bad inputs exit 1; both mutants abort; ctest same 5 failures as upstream main; pr_guard 0/0. Not postable as is: test_seq_rate.cpp has an unused include clang-tidy rejects (the trace's tidy never analysed the test). Post r3 = r2 minus that line plus body edits (dual wield limit, Checklist, README wording, Leap/Dragon Flight)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P17 (fid/seq-ias-r1) must pass an independent verifier before it is posted upstream. Done when: the verdict doc names the exact head to post. Consumer: the upstream PR queue (P17 row).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 r2 rebased on current upstream main, rebuilt, oracle and claim grids rerun
- [x] #2 game-side rule re-derived from game.exe and the grid axes judged
- [x] #3 mutants, full ctest vs upstream main, tidy, presubmit, pr_guard, trailers checked
- [x] #4 verdict and exact head in docs/upstream/verifier-p17-2026-11-15.md
<!-- AC:END -->
