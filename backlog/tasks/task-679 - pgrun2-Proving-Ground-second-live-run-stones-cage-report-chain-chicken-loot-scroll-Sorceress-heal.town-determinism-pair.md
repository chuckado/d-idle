---
id: TASK-679
title: >-
  pgrun2: Proving Ground second live run (stones, cage, report, chain, chicken,
  loot, scroll Sorceress, heal.town, determinism pair)
status: Done
assignee:
  - credit-runner
created_date: '2026-11-19T22:49:00Z'
updated_date: '2026-11-19T23:13:00Z'
labels:
  - docs
  - drill
dependencies: []
priority: medium
ordinal: 331000
card: "pgrun2"
owns:
  - "docs/probes/provingground-run-2026-11-19.md"
  - "docs/process/proving-ground.md"
  - "tools/provingground/scenarios/cain.json"
  - "tools/provingground/scenarios/behaviour-scenarios.json"
exclusive: []
contracts: []
result: "Verified live on the builder (lab slots, Paladin 11 and Sorceress 3, CHEATED setups, mechanics only): GREEN cain.report, heal.town (assert point fixed), loot.rare-ring-classified (with the points knob); RED findings (not fixed): cain.stones, cain.cage, cain.chain, chicken.recovery, cain.scroll Sorceress 3 (level goal plus the board start gate behind a fight, purse drain, banked points); cain.stones determinism pair same verdict and digest. Record: docs/probes/provingground-run-2026-11-19.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Run the not-yet-run Proving Ground scenarios; REDs are findings.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Scenarios run live on the builder and verdicts recorded
- [x] #2 heal.town assert timing fixed
- [x] #3 run record and coverage numbers updated
- [x] #4 stones determinism pair run
<!-- AC:END -->
