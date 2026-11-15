---
id: TASK-492
title: 'decomp26: P04 hit interrupts, P06 missile block gate, P09/P10/P03c issue drafts'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T11:51:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 274000
card: "decomp26"
owns:
  - "docs/research/decomp26-trace.md"
  - "docs/upstream/drafts/hitrec-interrupt.md"
  - "docs/upstream/bodies/hitrec-interrupt-body.md"
  - "docs/upstream/drafts/missile-block.md"
  - "docs/upstream/bodies/missile-block-body.md"
  - "docs/upstream/drafts/revive-mode-issue.md"
  - "docs/upstream/drafts/merc-aura-issue.md"
  - "docs/upstream/drafts/hitrec-knock-in-hit.md"
  - "docs/upstream/pipeline-2026-11-14.md"
  - "docs/upstream/pr-queue.md"
  - "docs/upstream/briefs/verify-hitrec-interrupt.md"
  - "docs/upstream/briefs/verify-missile-block.md"
  - "docs/research/data26/*"
exclusive:
  - "upstream-pr:hitrec-interrupt"
  - "upstream-pr:missile-block"
  - "fork-branch:fid/hitrec-interrupt-r1"
  - "fork-branch:fid/missile-block-r3"
result: "P04: fork branch fid/hitrec-interrupt-r1 (oracle 308e96f8, claim a78f66a3) on upstream fb6d799d: a hit interrupts a swing or cast whose skill interrupt column is set; V run both sides, 1,741 of 2,856 (skill row x mode) cells differ before, 0 after; ready for the independent verifier, post after #548. P06: fid/missile-block-r3 (oracle 13783a6d, claim cd314f2c): can-block = record physical over 0; 4,550 of 29,421 cases differ before (+42 of 45 skill-launch cases), 0 after; side finding d2d has no shafire2/4, unholybolt2/3 rows; ready for the verifier. P09 and P10 issue drafts and a P03c note written (P10: auras on the merc already merged upstream). Emulator on the worker host and the NAS builder only; verified live: no; nothing posted."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: keep the decomposition pipeline moving (P04, P06, P09, P10, P03c). Done when: both d2d deviations are proven by a differential with a fork branch for the verifier, and the issue drafts exist (docs/research/decomp26-trace.md).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 P04 oracle and claim built, V run before/after
- [x] #2 P06 re-check and build
- [x] #3 P09 P10 P03c drafts
<!-- AC:END -->
