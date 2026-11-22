---
id: TASK-626
title: 'verifygearabsorb3: verifier of the gear absorb r3 PR (fid/gear-absorb-r3)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T18:37:00Z'
updated_date: '2026-11-17T20:33:00Z'
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
result: "r3 FAIL (stale base f1533770, README oracle recipe breaks on log lines, ES/life-over-max/after-death untested: 4 mutants survive, merc reason wrong), r4 PASS WITH CHANGES: fid/gear-absorb-r4 ac5c3dd1+c05fa8e3 on upstream 895f8a67, pushed to the private fork; parent 9,639/9,519/9,484 (seeds 7, 5001, fresh 4242), claim 0; hand cases incl. life above max equal game.exe; 17/17 mutants abort; ctest 360/360 (two UI tests rerun alone); tidy 0; pr_guard 0/0; merc left out as scope (game.exe absorbs and heals it, I); body changes C1-C8; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: independent Opus check of fid/gear-absorb-r3 (TASK-621) before posting upstream. Consumer: the posting agent (docs/upstream/pr-queue.md). Done when: the verdict doc names the branch and commits to post and the body changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Rule re-derived from the decompilation; merc, Energy Shield, heal ordering and cap attacked
- [x] #2 Build, ctest, oracles, mutants, tidy, pr_guard, presubmit re-run on the builder
- [x] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
