---
id: TASK-443
title: 'survA: danger profiles (survival slice S-A)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-14T21:08:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 259000
card: "survA"
owns:
  - "internal/danger/"
  - "docs/survival/decisions-2026-11-14.md"
  - "docs/survival/danger.md"
  - "cmd/d2bot/dangerwire.go"
  - "cmd/d2bot/ladder.go"
  - "cmd/d2bot/live.go"
  - "cmd/d2bot/d2dbackend.go"
  - "internal/watch/funwire.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: slice S-A of docs/survival/design-2026-11-14.md (sections 3, 10, 12): the known quantity (zone danger) generated from the install tables, read-only, no behaviour change. Done when: the criteria below hold. Shared-file edits (ladder, live, d2dbackend, funwire) are one-line wiring of GET /api/danger.

result: "S-A shipped, verified on the Tower builder against the real install: Andariel life 1024 defence 60 level 12, Smith 404, Countess 96, Blood Raven 113 match boss1-trace; all 10 boss keys resolve against the real SuperUniques table (Build errors otherwise, TestShippedBossRows pins the rows); units per clear fitted per level type to areamon-table.csv (median 41 percent over 36 areas, maze levels worst, tolerance 45 percent); whole-repo vet, short tests, race on danger, watch, d2bot passed. /api/danger is read-only, no behaviour change. Not verified live in a running bot."
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/danger generator builds a profile per area and difficulty from the install tables
- [x] #2 encounters from group and density; bosses from boss1-trace data; override file with evidence check
- [x] #3 fixture and tests for two difficulties; real-install anchors (Blood Moor, Cold Plains, Black Marsh, Andariel); table diff against areamon-table.csv within tolerance
- [x] #4 GET /api/danger read-only, no behaviour change
- [x] #5 decisions-2026-11-14.md records the user answers
<!-- AC:END -->
