---
id: TASK-518
title: 'decomp33: upstream candidates P17 sequence rate, P18 Smite, P19 Holy Bolt heal'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-15T23:59:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 281000
card: "decomp33"
owns:
  - "docs/research/decomp33-p17-p19.md"
  - "tools/emu/dc33_seqrate.py"
  - "tools/emu/dc33_smite.py"
  - "tools/emu/dc33_holybolt.py"
  - "docs/upstream/bodies/seq-ias-body.md"
  - "docs/upstream/bodies/smite-range-body.md"
  - "docs/upstream/briefs/verify-seq-ias.md"
  - "docs/upstream/briefs/verify-smite-range.md"
  - "backlog/merge-lines/decomp33.txt"
exclusive:
  - "fork-branch:fid/seq-ias-r1"
  - "fork-branch:fid/smite-r1"
  - "fork-branch:fid/holybolt-heal-r1"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 P17: oracle then claim on the fork, compare 0 after, mutant aborts, checks
- [ ] #2 P18: re-derived against upstream main; branch or no-PR verdict
- [ ] #3 P19: re-derived against upstream main; branch or no-PR verdict
- [ ] #4 bodies, verifier briefs, pipeline and pr-queue rows, leakcheck clean
<!-- AC:END -->
