---
id: TASK-559
title: >-
  verifyseqfollow: independent verifier of the #565 follow-up
  fid/seq-ias-followup
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T18:54:00Z'
updated_date: '2026-11-16T19:43:00Z'
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
result: "PASS WITH CHANGES (NAS builder, not live). Diff test-only, identical to 310ff43d; nonexistent D2_PATCH_INSTALLER rc 1 both modes (b6e6940d: rc 0, ran on folder patch or skipped), valid installer rc 0 36 cells, unset/empty unchanged; compare 7600 same rc 0; test_game/test_combat rc 0; identity+Agent trailer, pr_guard 0, presubmit 0 FAIL, tidy 0. Body B1-B4 (before-behaviour wording, link #565 thread). Doc docs/upstream/verifier-seq-followup-2026-11-16.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Consumer: seqfollow (TASK-557). Why: independent check before posting the #565 follow-up. Done when: verdict and body changes written.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Verifier doc docs/upstream/verifier-seq-followup-2026-11-16.md with verdict and required body changes
- [x] #2 Failure path run on the builder: nonexistent D2_PATCH_INSTALLER rc 1, valid rc 0, unset unchanged
- [x] #3 Claim still proven: compare 7600 same, test_seq_rate 36 cells
<!-- AC:END -->
