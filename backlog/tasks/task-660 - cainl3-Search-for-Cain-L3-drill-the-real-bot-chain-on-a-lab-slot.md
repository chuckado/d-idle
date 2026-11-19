---
id: TASK-660
title: 'cainl3: Search for Cain L3 drill, the real bot chain on a lab slot'
status: Proposed
assignee:
  - live-agent
created_date: '2026-11-19T14:48:00Z'
labels:
  - d2d
  - drill
dependencies: []
priority: medium
ordinal: 326000
card: "cainl3"
owns:
  - "docs/probes/cainl3"
  - "docs/probes/search-for-cain-l3-2026-11-19.md"
  - "docs/probes/test-characters.md"
exclusive:
  - "d2d-control-socket"
contracts: []
result: "Verified live on a d2d lab slot (not Game.exe): the real bot chain ran unattended from the board tap to a finished quest (scroll, Akara read, stones, Tristram, archers, cage, return, quest word 0x2019) with a CHEATED clvl 32 Sorceress and combat.crowdAbandonN 0; four bot defects found (tree class -1 from npcs, finished chunks forgotten, hand-in tap taken in the field, free-Cain chunk done outside Tristram), fixed on agent/cainfix2. At the default crowd rule the hero chickens before the tree (open). Doc docs/probes/search-for-cain-l3-2026-11-19.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Real bot chain run unattended on a d2d lab slot, per-step timeline recorded
- [x] #2 Each failing step captured with log lines and a minimal repro
- [x] #3 Bot defects found are fixed on agent/cainfix2 with a must-fail test
<!-- AC:END -->
