---
id: TASK-628
title: 'verifycrush3: verifier of the Crushing Blow r3 PR (fid/crush-r3)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T00:49:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies: []
priority: high
ordinal: 315000
card: "verifycrush3"
owns:
  - "docs/upstream/verifier-crush3-2026-11-17.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M

Why: independent Opus check of fid/crush-r3 (crush3) before posting upstream. Consumer: the posting agent (docs/upstream/pr-queue.md). Done when: the verdict doc names the branch and commits to post and the body changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Rule re-derived from the decompilation; targets, divisor classes, ordering with leech and absorb, 256ths and resist edges attacked
- [ ] #2 Build, ctest, oracles on three seeds (one fresh), 8+ mutants, tidy, pr_guard, presubmit re-run on the builder
- [ ] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
