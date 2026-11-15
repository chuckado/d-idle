---
id: TASK-476
title: 'verifyp02: independent verifier of P02 fid/warp-walk-r2 (a clicked warp planned as a unit target)'
status: Proposed
assignee:
  - verifyp02
created_date: '2026-11-15T05:48:00Z'
labels:
  - docs
  - research
dependencies: []
priority: medium
ordinal: 266000
card: "verifyp02"
owns:
  - "docs/upstream/verifier-p02-2026-11-14.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: P02 (fid/warp-walk-r2) needs an independent adversarial verification before the user is asked for a go to post. Done when: the verifier doc gives a verdict with re-run counts and exact required changes. Consumer: decomp23 (P02 upstream PR).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 verifier doc with verdict, counts re-run, required changes
- [ ] #2 builds and oracle re-run on the NAS builder; test bites with the fix reverted
<!-- AC:END -->
