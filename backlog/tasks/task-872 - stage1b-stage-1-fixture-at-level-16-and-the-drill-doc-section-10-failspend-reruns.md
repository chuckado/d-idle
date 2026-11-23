---
id: TASK-872
title: 'stage1b: stage 1 fixture at level 16 and the drill doc section 10 (failspend reruns)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T08:07:00Z'
labels:
  - bot
  - drill
dependencies: []
priority: medium
ordinal: 393000
card: "stage1b"
owns:
  - "tools/provingground/scenarios/necro.json"
  - "docs/probes/necromancer-act1-drill-2026-11-22.md"
exclusive: []
contracts: []
result: "Freeze cause: spin without clock sleep after a refused step (fixed by failspend #1170). Stage 1 reruns RED at level 14 (plan walks into the Monastery, chunk starts in the Barracks) and RED at level 16 (Den hand-in refused after an 11 s clear, record 1 0x0004); Countess GREEN; Tools fix not live; Andariel unrun. Not ready."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Freeze cause and fix recorded
- [x] #2 Stage 1 rerun verdicts with defects named, nothing claimed GREEN
<!-- AC:END -->
