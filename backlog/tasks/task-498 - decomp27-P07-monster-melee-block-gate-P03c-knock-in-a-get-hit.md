---
id: TASK-498
title: 'decomp27: P07 monster melee block gate, P03c knock in a get-hit'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-15T14:52:00Z'
updated_date: '2026-11-15T15:08:00Z'
labels:
  - d2d
  - upstream-pr
dependencies: []
priority: medium
ordinal: 276000
card: "decomp27"
owns:
  - "docs/research/decomp27-trace.md"
  - "tools/emu/mel_block.py"
  - "docs/upstream/briefs/verify-knock-gethit.md"
  - "docs/upstream/briefs/verify-melee-block.md"
  - "docs/upstream/drafts/hitrec-knock-in-hit.md"
  - "docs/upstream/bodies/hitrec-knock-in-hit-body.md"
  - "docs/upstream/pr-queue.md"
  - "docs/upstream/pipeline-2026-11-14.md"
exclusive: []
contracts: []
result: "P07 melee block: d2d matches game.exe (FUN_0057ec10 can-block 1; 180 of 180 oracle cells, V run on the builder), no PR. P03c knock in a get-hit: proven (parent 1 of 13 cells differ, claim 0, V run), fork branch fid/knock-in-gethit-r1, ready for the Opus verifier. Not verified live; nothing posted."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Sonnet decomp27. Done when both items have a V run verdict and the verifier briefs exist.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 P07 oracle: game.exe FUN_0057ec10 vs d2d monster_blow, grid, verdict
- [x] #2 P03c oracle: FUN_0057edd0 new mode 0x13 vs d2d knock, modes 0..19, claim branch
- [x] #3 verifier briefs written, trace doc
<!-- AC:END -->
