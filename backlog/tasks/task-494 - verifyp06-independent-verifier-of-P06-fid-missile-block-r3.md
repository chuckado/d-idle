---
id: TASK-494
title: 'verifyp06: independent verifier of P06 fid/missile-block-r3'
status: In Progress
assignee:
  - verifyp06
created_date: '2026-11-15T12:25:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 275000
card: "verifyp06"
owns:
  - "docs/upstream/verifier-p06-2026-11-15.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P06 (monster missile block gate) needs an independent Opus verification before the user's push go. Consumer: decomp26 (P06 upstream PR). Done when: docs/upstream/verifier-p06-2026-11-15.md gives a verdict and the exact head to post.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Rebased rerun of the P06 claim with and without the fix on the NAS builder
- [ ] #2 Game-side oracle and case grid checked
- [ ] #3 PR hygiene checked and verdict with exact head written
<!-- AC:END -->
