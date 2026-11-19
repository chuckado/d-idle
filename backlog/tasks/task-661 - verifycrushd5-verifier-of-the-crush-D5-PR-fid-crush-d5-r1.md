---
id: TASK-661
title: 'verifycrushd5: verifier of the crush D5 PR (fid/crush-d5-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-19T14:49:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies: []
priority: high
ordinal: 327000
card: "verifycrushd5"
owns:
  - "docs/upstream/verifier-crushd5-2026-11-19.md"
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
- [ ] #1 Rule re-derived from the decompilation; boss classes mapped both sides; bit-2 writers scanned beyond the author's scan
- [ ] #2 Rebased on current upstream main; build, full ctest vs parent, oracle, 10+ mutants, tidy, pr_guard, presubmit on the builder
- [ ] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
