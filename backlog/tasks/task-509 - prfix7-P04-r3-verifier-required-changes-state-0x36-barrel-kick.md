---
id: TASK-509
title: 'prfix7: P04 r3 (verifier required changes: state 0x36, barrel kick)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T15:56:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 277000
card: "prfix7"
owns:
  - "docs/upstream/bodies/hitrec-interrupt-body.md"
  - "docs/upstream/drafts/hitrec-interrupt.md"
  - "docs/upstream/prfix7-p04-r3-2026-11-15.md"
  - "docs/upstream/briefs/verify2-hitrec-interrupt.md"
  - "docs/upstream/pr-queue.md"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive:
  - "upstream-pr:hitrec-interrupt"
  - "fork-branch:fid/hitrec-interrupt-r3"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 fid/hitrec-interrupt-r3 pushed to the private fork only (oracle first): state 0x36 and the barrel kick refuse
- [ ] #2 differential extended with the unit-state dimension: parent differs, claim 0, mutant aborts
- [ ] #3 body, draft, queue and pipeline rows updated; second verifier pass needed, nothing posted
<!-- AC:END -->
