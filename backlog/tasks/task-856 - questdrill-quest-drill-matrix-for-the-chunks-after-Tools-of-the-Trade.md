---
id: TASK-856
title: 'qdmatrix: quest drill matrix for the chunks after Tools of the Trade'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:09:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 387000
card: "qdmatrix"
owns:
  - "docs/probes/quest-drill-matrix-2026-11-23.md"
  - "tools/provingground/scenarios/questdrill.json"
  - "internal/power/archetype.go"
  - "internal/power/skilldamage_test.go"
  - "internal/questboard/state.go"
  - "internal/questboard/state_test.go"
  - "internal/nav/exit.go"
  - "internal/nav/exitgoal_test.go"
  - "internal/talk/talk.go"
  - "internal/talk/talk_test.go"
exclusive: []
contracts: []
result: "Board: a proxy done never beats an unmet prerequisite (Duriel and Hellforge showed done on an Act 1 hero). Nav: the walk to a gate has the gate as its goal (Sewers Level 2). Quest drill matrix started (docs/probes/quest-drill-matrix-2026-11-23.md). Three bot fixes with must-fail tests: power attack pick counts only points paid by clvl (no more unread attack for a clvl 19 Paladin), act-hop row unlocks on the boss reward pending, Warriv and Meshif hop row clicked while the act word reads needed. Sail east (Act 1 boat) GREEN Paladin 11 in the lab (CHEATED setup, not verified on the live hero)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Matrix doc lists every chunk through Act 2 with a verdict
- [ ] #2 Each RED has a root cause or a named blocker
<!-- AC:END -->
