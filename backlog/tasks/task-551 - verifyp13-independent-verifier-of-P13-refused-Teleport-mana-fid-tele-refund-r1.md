---
id: TASK-551
title: >-
  verifyp13: independent verifier of P13 refused Teleport mana
  (fid/tele-refund-r1)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-16T16:43:00Z'
updated_date: '2026-11-16T18:50:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: high
ordinal: 290000
card: "verifyp13"
owns:
  - "docs/upstream/verifier-p13-2026-11-16.md"
exclusive: []
result: "PASS WITH CHANGES (offline, NAS builder; not live). Rule re-derived (FUN_0056f7f0 0x56fa3d exit before FUN_0056bfe0; do 27 and FUN_00554ea0 return 0 only when nothing moved). r1 not postable: Claude identity and no Agent trailer, claim before oracle, compare exit 1 on its own claim (seed 2, own-subtile hops), tidy and presubmit failures. Post fid/tele-refund-r2 (oracle e7bfc411, claim 6bb8dade on upstream b6e6940d, pushed to the private fork): 124,000 cases, parent 57,001 refused hops DIFF, claim 0, mutant aborts; ctest as upstream bar a -j3 lobby flake. Body changes B1-B6 (only Teleport 0 and target-not-near are in the differential). Doc docs/upstream/verifier-p13-2026-11-16.md."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M
Consumer: TASK-547 (P13 upstream post).

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Verifier doc docs/upstream/verifier-p13-2026-11-16.md with verdict, required body changes and the exact branch and commits to post
- [x] #2 Game rule re-derived from the disassembly (FUN_0056f7f0, FUN_005ca360, FUN_00554ea0)
- [x] #3 Oracle and d2d side rebuilt and run on the NAS builder (parent, claim, mutant); compare exit codes checked
- [x] #4 Squashed oracle-first branch fid/tele-refund-r2 on the private fork remote
<!-- AC:END -->
