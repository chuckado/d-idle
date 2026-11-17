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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Verdict PASS / PASS WITH CHANGES / FAIL on fid/mana-burn-amount-r1 with the rule re-derived from the disassembly
- [ ] #2 Mutants (5+) tried against the oracle compare and the branch's tests; RNG reseed
- [ ] #3 Build and ctest on the builder (-j2), same failures as upstream; tidy, pr_guard, presubmit --quick
- [ ] #4 Fixed branch fid/mana-burn-amount-r2 on the private fork remote (clean identity, oracle-first) and the required body changes in docs/upstream/verifier-decomp38-2026-11-17.md
<!-- AC:END -->
