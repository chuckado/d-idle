---
id: TASK-608
title: 'verifydecomp38: independent verifier of the decomp38 Mana Burn amount PR candidate (fid/mana-burn-amount-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T10:29:00Z'
labels:
  - d2d
  - readonly
dependencies:
  - TASK-606
priority: high
ordinal: 308000
card: "verifydecomp38"
owns:
  - "docs/upstream/verifier-decomp38-2026-11-17.md"
exclusive: []
contracts: []
result: "r1 FAIL (no pinned test, oracle measured the field, presubmit 2 FAIL), r2 FAIL (tidy include), r3 PASS WITH CHANGES: fid/mana-burn-amount-r3 dc6d881a+177189d6 on upstream ed007569, pushed to the private fork; parent 1,750 / 1,600 of 2,000 outside game.exe's range, claim 0 / 0, rule exact 2,000 of 2,000 twice; ctest 356/356 (upstream 355/355); 8 mutants abort test_mana_burn; Energy Shield takes from the burn in game.exe (limit); Normal: only the unique burns; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Verdict PASS / PASS WITH CHANGES / FAIL on fid/mana-burn-amount-r1 with the rule re-derived from the disassembly
- [x] #2 Mutants (5+) tried against the oracle compare and the branch's tests; RNG reseed
- [x] #3 Build and ctest on the builder (-j2), same failures as upstream; tidy, pr_guard, presubmit --quick
- [x] #4 Fixed branch fid/mana-burn-amount-r2 on the private fork remote (clean identity, oracle-first) and the required body changes in docs/upstream/verifier-decomp38-2026-11-17.md
<!-- AC:END -->
