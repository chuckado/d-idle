---
id: TASK-772
title: 'boardsmrun: run the 18 quest-board SM scenarios live (Paladin 11, Sorceress 3)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T13:46:00Z'
updated_date: '2026-11-21T17:23:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 359000
card: "boardsmrun"
owns:
  - "tools/provingground/scenarios/board.json"
  - "cmd/d2scenario"
  - "internal/scenario"
  - "docs/probes/boardsm-scenarios-2026-11-21.md"
exclusive: []
contracts: []
result: "18 scenarios run live on the worker host (survtune-base d2d), Paladin 11 and Sorceress 3, CHEATED setups: 5 GREEN with determinism pair (now active), 8 RED (card 6 boardsm-api missing, potion-kit hold finding), 3 BLOCKED, 1 vacuous green; tooling fix for board samples. Record docs/probes/boardsm-scenarios-2026-11-21.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Scenarios flipped to active where they run; verdicts per scenario and class recorded in the probe doc
- [x] #2 Findings for RED results listed with a card name
- [x] #3 Determinism pair for every green
<!-- AC:END -->
