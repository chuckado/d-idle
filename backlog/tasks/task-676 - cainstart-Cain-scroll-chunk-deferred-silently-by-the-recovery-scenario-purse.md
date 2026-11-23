---
id: TASK-676
title: 'cainstart: Cain scroll chunk deferred silently by the recovery, scenario purse'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-19T20:03:00Z'
updated_date: '2026-11-19T20:26:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 330000
card: "cainstart"
owns:
  - "internal/bot/questboard.go"
  - "tools/provingground/scenarios/cain.json"
  - "docs/probes/cainstart-2026-11-19.md"
exclusive: []
contracts: []
result: "Verified on the builder (lab slot, CHEATED setup, mechanics only): cain.scroll Paladin/11 RED with purse 500 (broke-recovery defers the chunk silently), GREEN with purse 5000; the bot now shows the reason; Sorceress/3 still RED (open); sandbox variant not run. See docs/probes/cainstart-2026-11-19.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Gate decision shown and unit-tested
- [x] #2 cain.scroll Paladin seed 11 GREEN
<!-- AC:END -->
