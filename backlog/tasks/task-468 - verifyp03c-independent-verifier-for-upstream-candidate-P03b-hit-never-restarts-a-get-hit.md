---
id: TASK-468
title: 'verifyp03c: independent verifier for upstream candidate P03b (hit never restarts a get-hit)'
status: Proposed
assignee:
  - verifyp03c
created_date: '2026-11-15T04:41:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 265000
card: "verifyp03c"
owns:
  - "docs/upstream/verifier-p03b-2026-11-14.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: upstream PR candidate P03b (fork branch fid/hitrec-norestart-r2) needs an independent check before posting. Done when: the verifier doc gives PASS / PASS WITH CHANGES / FAIL with exact required changes. Consumer: the upstream PR pipeline (P03b).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Verifier doc with verdict and required changes
- [ ] #2 Builds and tests rerun on the NAS builder, mutant check
- [ ] #3 Emulator oracle rerun and FUN_0057edd0 read
- [ ] #4 Stacking with and without #541 checked
- [ ] #5 PR hygiene checked
<!-- AC:END -->
