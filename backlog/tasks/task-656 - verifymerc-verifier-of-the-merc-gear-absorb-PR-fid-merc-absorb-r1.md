---
id: TASK-656
title: 'verifymerc: verifier of the merc gear absorb PR (fid/merc-absorb-r1)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T23:23:00Z'
labels:
  - d2d
  - readonly
  - d2d-combat
  - fidelity
dependencies: []
priority: high
ordinal: 246000
card: "verifymerc"
owns:
  - "docs/upstream/verifier-mercabsorb-2026-11-18.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Why: independent Opus check of fid/merc-absorb-r1 (decomp41) before posting upstream. Consumer: the posting agent (docs/upstream/pr-queue.md). Done when: the verdict doc names the branch and commits to post and the body changes.

Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Rule re-derived; merc gear stat path (dead code?), pets and summons, heal ordering attacked
- [ ] #2 Build, ctest against the parent, oracles on three seeds (one fresh), 8+ mutants, tidy, pr_guard, presubmit re-run on the builder
- [ ] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
