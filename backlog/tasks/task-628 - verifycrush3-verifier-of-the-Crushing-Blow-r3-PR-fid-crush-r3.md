---
id: TASK-628
title: 'verifycrush3: verifier of the Crushing Blow r3 PR (fid/crush-r3)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-18T00:49:00Z'
updated_date: '2026-11-18T01:46:00Z'
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
result: "r3 FAIL (8 of 18 mutants pass test_combat, m11 plain-multiply resist caught by nothing; RE finding missing, monsters.md says after the hit; divisor row says unique where game.exe mask 2 is superunique), r4 PASS WITH CHANGES: fid/crush-r4 7476c90c+760df1fa+1da9632a on upstream bebfdbcf, pushed to the private fork; parent 2,223/2,243/2,234 crush and 2,745/2,746/2,745 pair (seeds 7, 5001, fresh 9091), claim 0; 18/18 mutants abort r4 test_combat; ctest 369/370 (test_stand_attack fails on bebfdbcf alone); tidy 0; pr_guard 0/0; presubmit 0 FAIL 3 WARN; body changes C1-C9; follow-ups D5 unique divisor, D6 get-hit counts the crush; offline, not verified live"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M

Why: independent Opus check of fid/crush-r3 (crush3) before posting upstream. Consumer: the posting agent (docs/upstream/pr-queue.md). Done when: the verdict doc names the branch and commits to post and the body changes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Rule re-derived from the decompilation; targets, divisor classes, ordering with leech and absorb, 256ths and resist edges attacked
- [x] #2 Build, ctest, oracles on three seeds (one fresh), 8+ mutants, tidy, pr_guard, presubmit re-run on the builder
- [x] #3 Verdict doc with required body changes and the exact branch and commits (pushed to the private fork)
<!-- AC:END -->
