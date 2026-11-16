---
id: TASK-559
title: 'verifyseqfollow: independent verifier of the #565 follow-up fid/seq-ias-followup'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T18:54:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 293000
card: "verifyseqfollow"
owns:
  - "docs/upstream/verifier-seq-followup-2026-11-16.md"
exclusive: []
contracts: []
result: "INTERIM: read checks pass; builder runs not done (OOM). Provisional PASS WITH CHANGES, body B1-B4. Not live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Consumer: seqfollow (TASK-557). Why: independent check before posting the #565 follow-up. Done when: verdict and body changes written.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Verifier doc docs/upstream/verifier-seq-followup-2026-11-16.md with verdict and required body changes
- [ ] #2 Failure path run on the builder: nonexistent D2_PATCH_INSTALLER rc 1, valid rc 0, unset unchanged
- [ ] #3 Claim still proven: compare 7600 same, test_seq_rate 36 cells
<!-- AC:END -->
