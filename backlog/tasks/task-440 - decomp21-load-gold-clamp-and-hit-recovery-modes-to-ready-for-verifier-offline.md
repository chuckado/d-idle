---
id: TASK-440
title: 'decomp21: load-gold clamp and hit-recovery modes to ready-for-verifier (offline)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T21:01:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 258000
card: "decomp21"
owns:
  - "docs/upstream/drafts/load-gold-clamp.md"
  - "docs/upstream/drafts/hitrec-modes.md"
  - "docs/upstream/offline-decomp21-2026-11-14.md"
  - "docs/upstream/bodies/load-gold-clamp-body.md"
  - "docs/upstream/bodies/hitrec-modes-body.md"
exclusive:
  - "fork-branch:fid/load-gold-clamp-r2"
  - "fork-branch:fid/hitrec-modes-r1"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Offline on the NAS builder; no emulator, no game. Done when both items are ready for the verifier.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 P05 load-gold clamp: fork branch split into oracle and claim, rebased on upstream main, d2d-side before/after counted offline, draft written
- [ ] #2 P03 hit-recovery modes: fork branch with oracle and two claims, d2d-side before/after counted offline, draft written
- [ ] #3 Pipeline and pr-queue rows updated; blockers (emulator game-side rerun, verifier) named
<!-- AC:END -->
