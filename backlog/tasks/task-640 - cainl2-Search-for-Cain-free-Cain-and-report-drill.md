---
id: TASK-640
title: 'cainl2: Search for Cain free-Cain and report drill'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-18T15:16:00Z'
labels:
  - d2d
  - drill
dependencies: []
priority: medium
ordinal: 321000
card: "cainl2"
owns:
  - "docs/probes/cainl2"
  - "docs/probes/search-for-cain-l2-2026-11-18.md"
  - "docs/probes/test-characters.md"
exclusive:
  - "d2d-control-socket"
contracts: []
result: "Verified live on a d2d lab slot (not Game.exe). Cage = class 26 object at (27.3, 27.7) in Tristram, walk to ~(25,27) then interact (state NU to ON, quest record 4 0x0010 to 0x2012, Cain walks off alone); TP scroll and portal back work; talking to Akara sets record 4 to 0x2011; three champion archers at the arrival are lethal at level 4. Doc docs/probes/search-for-cain-l2-2026-11-18.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Cage location, action sequence and quest flags recorded with fixtures
- [x] #2 Return path and the report to Akara step recorded
- [x] #3 CAIN-5b inputs written
<!-- AC:END -->
