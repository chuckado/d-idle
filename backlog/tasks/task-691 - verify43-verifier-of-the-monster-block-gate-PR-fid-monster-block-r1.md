---
id: TASK-691
title: 'verify43: verifier of the monster block gate PR (fid/monster-block-r1)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-20T18:24:00Z'
updated_date: '2026-11-20T19:32:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies: []
priority: high
ordinal: 336000
card: "verify43"
owns:
  - "docs/research/decomp43-monster-block-verify.md"
exclusive: []
contracts: []
result: "GO WITH FIXES for fid/monster-block-r1: FUN_00622720/FUN_006225f0 re-disassembled (jump table decoded: classes 243/310/333 forced open, 359 closed, no outcome change); fresh claim run 0 of 2,469 differ, parent 228, compare exits 1 on 14 bad-input cases; all 17 non-equivalent mutants abort the pins (m10-m13 rerun as ai.hpp edits after they failed to compile in job 20261119-225204-80bf; m14, m18 equivalent); fixes: rebase (ai.hpp conflict with upstream gear_res), test default mode must return D2D_TEST_SKIPPED, body commit count and mutant paragraph; new deviation D6 player pets never block (I); #634 and #637 merged, no human comment; game-side rerun byte-identical, MB_SPOT=5 room control 0 differ, codes and cap reproduce (job 20261120-132422-41a1); offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Consumer: the coordinator's upstream decision on fid/monster-block-r1 (decomp43, TASK-683). Why: the independent Opus verifier is required before any upstream post. Done when: the verify doc has a verdict and the evidence.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Claim re-derived from the oracle and the decomp trace; compare exit codes checked on bad input
- [x] #2 Remaining mutants run on the builder long lane with results recorded
- [x] #3 Commit hygiene, CI/CodeRabbit fit and side-effect risks reviewed; open upstream PRs #634 and #637 checked read-only
- [x] #4 Verdict GO / GO WITH FIXES / NO-GO written in docs/research/decomp43-monster-block-verify.md
<!-- AC:END -->
