---
id: TASK-483
title: 'prfix5: P02 r3b, empty path not equal, repeatability'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T08:30:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 270000
card: "prfix5"
owns:
  - "docs/upstream/bodies/warp-walk-plan-body.md"
  - "docs/upstream/drafts/warp-walk-plan.md"
  - "docs/upstream/pr-queue.md"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive:
  - "fork-branch:fid/warp-walk-r3"
contracts: []
result: "fid/warp-walk-r3 now 5464877b (pushed to the private fork only; previous head kept as fid/warp-walk-r3-pre2): an empty tick path is not a match in compare and the test (plan equal in 60,747 compared starts, 0 differ, 427 not compared; test 8 of 8, mutant 6 of 8 fail); counts are not repeatable across processes (same binary, 13 of 46,000 lines differ, source U), wide residue set by key; verified on the NAS builder and emulator, not live; verifier to recheck these two items."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Why: the second verifier pass of P02 r3 (verifier2-p02-2026-11-15) found an empty tick path counted as equal and zero headroom in the counts. Done when: both fixed on fid/warp-walk-r3, rerun, body updated.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 compare and the test treat an empty tick path as not compared, rerun on the NAS builder
- [x] #2 body updated, repeatability finding written
<!-- AC:END -->
