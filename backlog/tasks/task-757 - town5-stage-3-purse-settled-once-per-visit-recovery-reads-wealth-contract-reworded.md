---
id: TASK-757
title: 'town5: stage 3 purse settled once per visit, recovery reads wealth, contract reworded'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:51:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-715
priority: medium
ordinal: 345000
card: "town5"
owns:
  - "internal/town"
  - "internal/resource"
  - "town-basics"
exclusive: []
contracts:
  - "town-basics"
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 one stash stop settles the purse to planned spend + chunk gold + carry line (300+50*clvl); recovery/poverty read wealth, unreadable stash counts 0 and is logged
- [ ] #2 contract line 'Bank before the shop' reworded to 'the purse is settled once per visit to planned spend + chunk gold + carry line' and the sentence shown in the result
- [ ] #3 scenarios broke-after-bank, malus-errand-gold
<!-- AC:END -->
