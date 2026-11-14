---
id: TASK-418
title: 'prverify6: second verifier pass on swing-start r2 (fid/swing-start-r2)'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-14T15:12:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 254000
card: "prverify6"
owns:
  - "docs/upstream/verification-swing-start-r2-2026-11-14.md"
exclusive: []
contracts: []
result: "NEEDS-WORK, text only, on r2 (not verified live; offline differential V run on the worker host): five first-pass edits correct, held-weapon test fails on r1 and passes on r2, new 78,172-cell real-path grid 0 of 51,056 reachable cells differ on r2 (parent 18,064, r1 5,940), class-restricted exclusion legitimate; required: oracle commit message 12,460/2,101 -> 11,270/2,010, drop the stale worn-or-not sentence in the draft; docs/upstream/verification-swing-start-r2-2026-11-14.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user's rule needs an independent verifier PASS before the swing-start PR (#680, queue row 1c) is posted. Consumer: TASK-417 (prswing2). Done when: the r2 verdict doc is on this branch.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Edits 1 to 5 of the first pass checked on r2 (code, test, counts, wording)
- [x] #2 New-grid differential through d2d's real attack path, parent vs r2, exclusions checked with d2d's equip rules
- [x] #3 Regressions, builds, ctest, presubmit, pr_guard, hygiene and upstream state checked
- [x] #4 Verdict doc written: docs/upstream/verification-swing-start-r2-2026-11-14.md
<!-- AC:END -->
