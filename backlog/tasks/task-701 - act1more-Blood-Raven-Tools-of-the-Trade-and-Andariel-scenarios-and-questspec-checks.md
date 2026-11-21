---
id: TASK-701
title: 'act1more: Blood Raven, Tools of the Trade and Andariel scenarios and questspec checks'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-20T23:48:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 338000
card: "act1more"
owns:
  - "tools/provingground/scenarios/act1more.json"
  - "internal/questspec/act1boss_test.go"
  - "docs/probes/provingground-act1more-2026-11-20.md"
result: "Blood Raven GREEN on the Proving Ground (Paladin/11, job 20261120-195436-d0c5, determinism pair with 20261120-210945-74a3: same verdict, digest differs; CHEATED setup, mechanics only, Kashya hand-in U). Tools of the Trade BLOCKED (route guard wants 4 healing potions, never restocked from the pack), Andariel RED (no Start offered at clvl 26, Outer Cloister nav); Sorceress/3 not run; findings and card list in docs/probes/provingground-act1more-2026-11-20.md. questspec boss checks green, -race -count=20. Not verified on a played hero."
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
- [x] #1 Scenarios act1.burial.blood_raven, act1.tools.malus, act1.andariel.slaughter written
- [x] #2 questspec data checks for the three boss chains and the Blood Raven title
- [x] #3 Each run on the builder, GREEN/RED/BLOCKED recorded with stall lines, determinism pair for any green
- [x] #4 Findings and cards listed in the probe doc
<!-- AC:END -->
