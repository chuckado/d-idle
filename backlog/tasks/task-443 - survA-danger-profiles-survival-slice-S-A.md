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
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 internal/danger generator builds a profile per area and difficulty from the install tables
- [ ] #2 encounters from group and density; bosses from boss1-trace data; override file with evidence check
- [ ] #3 fixture and tests for two difficulties; real-install anchors (Blood Moor, Cold Plains, Black Marsh, Andariel); table diff against areamon-table.csv within tolerance
- [ ] #4 GET /api/danger read-only, no behaviour change
- [ ] #5 decisions-2026-11-14.md records the user answers
<!-- AC:END -->
