---
id: TASK-648
title: 'cairnportal: PR candidate D1, the Cairn Stones'' Tristram portal offset (fid/cairn-portal-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T17:04:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 323000
card: "cairnportal"
owns:
  - "docs/upstream/cairn-portal-2026-11-18.md"
  - "docs/upstream/bodies/cairn-portal-body.md"
exclusive: []
contracts: []
result: "PR candidate D1 built, NOT posted: fid/cairn-portal-r1 on upstream 52d4cdd1, oracle 463e440b + claim f8e4ac32 + docs f81ca8da, pushed to the private fork; game.exe FUN_00593710/FUN_005a9930/FUN_00592d50 run natively: parent 39 of 39 rows differ on the stones portal and the timer portal (d2d offset 5x), claim 0; 12 of 12 mutants fail test_cairn_portal; ctest 393/394 (smoke_d2d fails on upstream main too); tidy 0, pr_guard 0/0, presubmit 0 FAIL; Gibbet fallback line not touched (U); offline, not verified live; awaiting the independent verifier"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: D1 of docs/research/decomp40-cairn-portal.md (the portal offset added in tiles, 5x too far) as an upstream PR candidate with an oracle that runs game.exe. Consumer: the independent verifier (d2-reviewer) and the posting agent (docs/upstream/pr-queue.md). Done when: the branch is on the private fork, the body and trace doc are in this repo, nothing is posted.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Oracle commit: game.exe's FUN_00593710 / FUN_00592d50 run natively, d2d dump, compare exits 1 on parent
- [x] #2 Claim commit: the fix plus pins; mutants abort; ctest same failures as upstream; tidy, pr_guard, presubmit
- [x] #3 Branch fid/cairn-portal-r1 pushed to the private fork; body and trace doc in this repo; not posted
<!-- AC:END -->
