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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user's rule needs an independent verifier PASS before the swing-start PR (#680, queue row 1c) is posted. Consumer: TASK-417 (prswing2). Done when: the r2 verdict doc is on this branch.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Edits 1 to 5 of the first pass checked on r2 (code, test, counts, wording)
- [ ] #2 New-grid differential through d2d's real attack path, parent vs r2, exclusions checked with d2d's equip rules
- [ ] #3 Regressions, builds, ctest, presubmit, pr_guard, hygiene and upstream state checked
- [ ] #4 Verdict doc written: docs/upstream/verification-swing-start-r2-2026-11-14.md
<!-- AC:END -->
