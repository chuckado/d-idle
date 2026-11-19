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
result: "r1 FAIL (stale base 9b64c11a; heal cap at the merc's whole max life unpinned: mutant v3 passes r1's pins and every oracle case; combat.md says FUN_0057c6c0 tests no unit type; body gaps), r2 PASS WITH CHANGES: fid/merc-absorb-r2 e724e726+fc7818e9 on upstream 563f9257, pushed to the private fork; not dead code (merc equip, merc_gear, make_fighter); parent 3,075/3,056/3,057/3,077 (seeds 7, 5001, 9091, fresh 4242), claim 0; player = merc 300/300; 14/14 mutants abort; ctest 400/401 (smoke_d2d fails on upstream too); tidy 0; pr_guard 0/0; presubmit 0 FAIL 5 WARN; body changes C1-C8; equip path in game.exe stays inferred; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: independent Opus check of fid/merc-absorb-r1 (decomp41) before posting upstream. Consumer: the posting agent (docs/upstream/pr-queue.md). Done when: the verdict doc names the branch and commits to post and the body changes.

Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Rule re-derived; merc gear stat path (dead code?), pets and summons, heal ordering attacked
- [x] #2 Build, ctest against the parent, oracles on three seeds (one fresh), 8+ mutants, tidy, pr_guard, presubmit re-run on the builder
- [x] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
