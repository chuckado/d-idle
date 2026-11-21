---
id: TASK-704
title: 'tphoard: loose town portal scrolls hoard the pack beside a tome'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T00:26:00Z'
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
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Sandbox hero: a tome of 10 plus 22 loose TP scrolls, pack 82 percent. Done when the loose count stays at the cap.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 pickup rule caps loose scrolls beside a tome (town.looseScrollKeep), two-class test
- [ ] #2 sell chore sells the surplus loose scrolls, two-class test
- [ ] #3 scenario town.tp-scroll-hoard RED before, GREEN after on the builder
- [ ] #4 knobs doc, contracts and backlog checks, leakcheck clean
<!-- AC:END -->
