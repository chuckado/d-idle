---
id: TASK-646
title: >-
  decomp40: Cairn Stones order, Tristram portal and the scroll drop, game.exe vs
  d2d
status: Done
assignee:
  - credit-runner
created_date: '2026-11-18T16:40:00Z'
updated_date: '2026-11-18T17:01:00Z'
labels:
  - d2d
  - research
dependencies: []
priority: medium
ordinal: 246000
card: "decomp40"
owns:
  - "docs/research/decomp40-cairn-portal.md"
exclusive: []
contracts: []
result: "Offline, not verified live, no V run. d2d applies the Tristram portal offset (+6,-3) in tiles instead of subtiles (D1, V read + lab D): portal 5x too far, outside the ring; d2d never sends the scroll order (D2). Stone rule equals game.exe; red portal already in upstream main. PR candidate D1 plan only. docs/research/decomp40-cairn-portal.md"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: a d2d contributor reported the portal outside the pillars and the scroll order not shown. Done when: the trace is committed with tags. Consumer: cain5b / CAIN-4,5.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Trace doc decomp40-cairn-portal.md answers Q1 to Q3 with addresses and tags
- [x] #2 Deviation list, bot findings, PR candidate verdict and drill recipe written
<!-- AC:END -->
