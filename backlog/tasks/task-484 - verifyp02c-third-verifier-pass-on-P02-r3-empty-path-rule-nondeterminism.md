---
id: TASK-484
title: 'verifyp02c: third verifier pass on P02 r3 (empty-path rule, nondeterminism)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T08:35:00Z'
updated_date: '2026-11-15T09:27:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 270000
card: "verifyp02c"
owns:
  - "docs/upstream/verifier3-p02-2026-11-15.md"
exclusive: []
result: "P02 r3 head 5464877b PASS: empty tick path no longer a match (426 not-compared starts all have d2d unit plan = game.exe); claim 0 plan differ on every run incl. a deterministic one. Nondeterminism source: async level builds (GameData::want_level std::async) link neighbours late, so edge bits vary; prebuilding all levels makes runs byte-identical; ASan/UBSan clean. Recommended: prebuild line in test_warp_walk (+ known pair 40:2). Offline, not live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: recheck of P02 r3 head 5464877b after verifier 2. Consumer: prfix5. Done when: the doc gives a verdict and the nondeterminism source or a bound.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 docs/upstream/verifier3-p02-2026-11-15.md with verdict and the nondeterminism source or bound
<!-- AC:END -->
