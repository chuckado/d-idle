---
id: TASK-621
title: 'gearabsorb3: gear absorb r3 (heal-back, life oracle)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T15:03:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 310000
card: "gearabsorb3"
owns:
  - "docs/upstream/bodies/gear-absorb-body.md"
  - "docs/research/gear-absorb-r3.md"
exclusive: []
contracts: []
result: "fid/gear-absorb-r3 on the private fork remote (oracle 982a65c1, claim 270b9799, upstream f1533770): the heal-back per blow, a life oracle (parent 9,639 and 9,519 of 20,000 differ, claim 0), players only (merc left out on purpose), 12 of 12 mutants abort, ctest 359 of 359 (smoke_d2d once under load). Offline, nothing posted, no live game; verifier-ready (Opus re-verification next)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the verifier failed gear absorb r2 (the heal-back missing; docs/upstream/verifier-decomp37-2026-11-17.md A1 to A6). Done when: fid/gear-absorb-r3 is built, tested and pushed to the private fork, the body and the trace are written, and the re-verification can start.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 r3 fork branch fid/gear-absorb-r3 built, tests, mutants
- [x] #2 body gear-absorb-body.md per A1-A6
- [x] #3 trace doc docs/research/gear-absorb-r3.md
<!-- AC:END -->
