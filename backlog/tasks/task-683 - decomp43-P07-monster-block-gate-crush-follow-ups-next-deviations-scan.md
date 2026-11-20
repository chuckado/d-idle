---
id: TASK-683
title: 'decomp43: P07 monster block gate, crush follow-ups, next deviations scan'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T03:12:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 246000
card: "decomp43"
owns:
  - "docs/research/decomp43-monster-block-trace.md"
  - "docs/upstream/bodies/monster-block-body.md"
  - "tools/emu/mb_explore.py"
  - "tools/emu/mb_gate.py"
  - "tools/emu/mb_act2.py"
  - "tools/emu/mb_col.py"
  - "tools/emu/mb_block.py"
  - "tools/emu/mb_mut.py"
exclusive:
  - "fork-branch:fid/monster-block-r1"
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Consumer: the independent verifier of fid/monster-block-r1 (the coordinator schedules it).

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 trace doc: P07 monster block rule with addresses and tags, d2d and bot compared
- [ ] #2 oracle differential of the block chance against game.exe (V run), parent mismatches, claim 0
- [ ] #3 mutants and pins of the claim
- [ ] #4 crush follow-up (merc, pet, missile) decided
- [ ] #5 scan of the next player-visible deviations
<!-- AC:END -->
