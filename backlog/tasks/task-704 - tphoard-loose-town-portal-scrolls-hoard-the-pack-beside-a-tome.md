---
id: TASK-704
title: 'tphoard: loose town portal scrolls hoard the pack beside a tome'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T00:26:00Z'
updated_date: '2026-11-21T02:30:00Z'
labels:
  - bot
  - bug
dependencies: []
priority: medium
ordinal: 338000
card: "tphoard"
owns:
  - "internal/loot/loot.go"
  - "internal/loot/lootskip_test.go"
  - "internal/town/shop.go"
  - "internal/town/spare_test.go"
  - "internal/env/tuning_town.go"
  - "internal/scenario/state.go"
  - "internal/scenario/newkinds_test.go"
  - "tools/provingground/scenarios/tphoard.json"
  - "tools/provingground/scenarios/behaviours.json"
  - "docs/knobs.md"
  - "docs/tphoard.md"
  - "internal/town/trip.go"
  - "internal/town/staleportal_test.go"
exclusive: []
contracts: []
result: "Loose TP/ID scrolls beside a tome capped by town.looseScrollKeep (2) in the pickup rule and sold down by the sell chore; the in-town fresh-portal cast (cause of the failed casts) removed. Scenario town.tp-scroll-hoard RED (control knob 99) then GREEN on the builder lab slots (Paladin, Sorceress), CHEATED setup, mechanics only; not verified on the sandbox hero."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Sandbox hero: a tome of 10 plus 22 loose TP scrolls, pack 82 percent. Done when the loose count stays at the cap.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 pickup rule caps loose scrolls beside a tome (town.looseScrollKeep), two-class test
- [x] #2 sell chore sells the surplus loose scrolls, two-class test
- [x] #3 scenario town.tp-scroll-hoard RED before, GREEN after on the builder
- [x] #4 knobs doc, contracts and backlog checks, leakcheck clean
<!-- AC:END -->
