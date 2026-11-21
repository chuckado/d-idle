---
id: TASK-496
title: 'verifyp06b: recheck of P06 fid/missile-block-r4'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-15T14:10:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 276000
card: "verifyp06b"
owns:
  - "docs/upstream/verifier2-p06-2026-11-15.md"
exclusive: []
contracts: []
result: "P06 r4 recheck: PASS, post fid/missile-block-r4 3ca8ee32 (not verified live; NAS builder + one emulator launch run). All 5 required and 3 recommended fixes in; maul game record 2,560-5,120 can-block 1; old ordering 15 differ and aborts test_carried_missile. docs/upstream/verifier2-p06-2026-11-15.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: recheck of the P06 r4 head after verifier-p06. Consumer: decomp26 (P06 upstream PR). Done when: docs/upstream/verifier2-p06-2026-11-15.md gives a verdict and the head to post.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 r4 rebuilt and rerun on the NAS builder: oracle, skill compare, carried test with the old ordering, ctest
- [x] #2 All five required and three recommended fixes of verifier-p06 checked
- [x] #3 Verdict and exact head written
<!-- AC:END -->
