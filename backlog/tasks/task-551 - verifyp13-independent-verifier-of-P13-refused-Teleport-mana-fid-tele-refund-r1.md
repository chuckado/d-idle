---
id: TASK-551
title: 'verifyp13: independent verifier of P13 refused Teleport mana (fid/tele-refund-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T16:43:00Z'
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
- [ ] #1 Verifier doc docs/upstream/verifier-p13-2026-11-16.md with verdict, required body changes and the exact branch and commits to post
- [ ] #2 Game rule re-derived from the disassembly (FUN_0056f7f0, FUN_005ca360, FUN_00554ea0)
- [ ] #3 Oracle and d2d side rebuilt and run on the NAS builder (parent, claim, mutant); compare exit codes checked
- [ ] #4 Squashed oracle-first branch fid/tele-refund-r2 on the private fork remote
<!-- AC:END -->
