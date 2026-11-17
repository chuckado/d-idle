---
id: TASK-626
title: 'verifygearabsorb3: verifier of the gear absorb r3 PR (fid/gear-absorb-r3)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T18:37:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies:
  - TASK-621
priority: high
ordinal: 313000
card: "verifygearabsorb3"
owns:
  - "docs/upstream/verifier-gearabsorb3-2026-11-17.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: independent Opus check of fid/gear-absorb-r3 (TASK-621) before posting upstream. Consumer: the posting agent (docs/upstream/pr-queue.md). Done when: the verdict doc names the branch and commits to post and the body changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Rule re-derived from the decompilation; merc, Energy Shield, heal ordering and cap attacked
- [ ] #2 Build, ctest, oracles, mutants, tidy, pr_guard, presubmit re-run on the builder
- [ ] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
