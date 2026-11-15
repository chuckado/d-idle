---
id: TASK-495
title: 'prfix6: P06 missile block r4 (verifier required changes)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-15T13:17:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 275000
card: "prfix6"
owns:
  - "docs/upstream/bodies/missile-block-body.md"
  - "docs/upstream/drafts/missile-block.md"
  - "docs/upstream/pr-queue.md"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive:
  - "upstream-pr:missile-block"
  - "fork-branch:fid/missile-block-r4"
contracts: []
result: "P06 r4: fork branch fid/missile-block-r4 pushed (private fork remote only): oracle 6d4b9971, claim 3ca8ee32 on upstream aaa17aee. Verifier changes applied: unblockable set after the carried skill physical (Death Mauler maul stays blockable; game side V run can-block 1 on its launch record; the earlier ordering differs in 15 maul cases and aborts test_carried_missile), EMU no longer required, real Agent trailers, Checklist and siblings line, cstdint and test includes for clang-tidy, limits lines. Oracle compare 29,421 cases 0 differ (parent 4,550), 60 skill cases 0 differ (parent 42); ctest same two failures as upstream; mutant aborts. A second verifier pass is needed before posting; nothing posted. Verified live: no."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Death Mauler ordering fix with test, oracle rerun 0 differ
- [x] #2 r4 branch built, tested, pushed to the fork remote only
- [x] #3 body, draft, queue rows updated; not posted
<!-- AC:END -->
