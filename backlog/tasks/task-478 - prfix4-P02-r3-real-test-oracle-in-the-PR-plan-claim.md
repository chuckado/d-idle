---
id: TASK-478
title: 'prfix4: P02 r3, real test, oracle in the PR, plan claim'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T06:34:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 268000
card: "prfix4"
owns:
  - "docs/upstream/bodies/warp-walk-plan-body.md"
  - "docs/upstream/drafts/warp-walk-plan.md"
  - "docs/upstream/pr-queue.md"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive:
  - "fork-branch:fid/warp-walk-r3"
contracts: []
result: "Fork branch fid/warp-walk-r3 (d442f595, pushed to the private fork only; r2 kept as fid/warp-walk-r2-pre-squash): test fails without the fix (6 of 8 starts), compare exits 1 on the parent and 0 on the claim, plan = game.exe in 60,955 of 61,206 starts (parent 981 differ), outcome 146 to 0 and 211 to 157 (157 outside the plan, U); ctest equals the parent apart from the new test; body written; verified on the NAS builder and the emulator, not live; second verifier pass needed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the verifier of P02 (verifyp02) asked for a test that fails without the fix, an oracle and `compare` in the PR that tell parent from claim, a plan-only claim and clean commits. Done when: fork branch `fid/warp-walk-r3` is built and rerun, the PR body is written, the queue rows are updated.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fork branch fid/warp-walk-r3 built, test fails with the fix reverted, compare fails parent and passes claim
- [x] #2 PR body and draft written, queue rows updated
<!-- AC:END -->
