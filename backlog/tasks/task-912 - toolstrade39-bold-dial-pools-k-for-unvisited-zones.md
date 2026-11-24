---
id: TASK-912
title: 'toolstrade39: bold dial pools k for unvisited zones'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T02:45:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 417000
card: "toolstrade39"
owns:
  - "internal/survival/controller.go"
  - "internal/survival/params.go"
  - "internal/survival/params.json"
  - "internal/survival/pooledk_test.go"
  - "docs/probes/zone-pooledk.md"
exclusive: []
contracts: []
result: "Cellar rows are rejected by margin: the hero k sits at its clamp 3 and every zone without kills carries it whole, visited fields have k 1.5; xp never enters eligibility. Bold dial now pools the k of visited zones for unvisited ones (poolUnvisitedK). Offline tests only, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Real cause of the Cellar rows being rejected documented with the live line
- [x] #2 poolUnvisitedK flag (bold only) with a must-fail test on two builds
<!-- AC:END -->
