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
result: "P04 r3: fork branch fid/hitrec-interrupt-r3 pushed (private fork remote only): oracle 113ab0d7, claim 5fa67510 on upstream 77990456. The first verifier required changes applied: state 0x36 (Leap Attack in flight) and the barrel kick (Kick) refuse; the oracle has a unit-state dimension. Compare 5,711 cells and the barrel kick: parent 2,811 differ, r2 claim 2,681 (2,680 state 0x36 and the barrel), claim 0; both mutants abort; selftest 16 of 16; ctest same two failures as upstream; pr_guard 0/0, presubmit --quick 0 FAIL. Concentration aura (20 percent) stated as a limit. A second verifier pass is needed before posting; nothing posted. Verified live: no."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 fid/hitrec-interrupt-r3 pushed to the private fork only (oracle first): state 0x36 and the barrel kick refuse
- [x] #2 differential extended with the unit-state dimension: parent differs, claim 0, mutant aborts
- [x] #3 body, draft, queue and pipeline rows updated; second verifier pass needed, nothing posted
<!-- AC:END -->
