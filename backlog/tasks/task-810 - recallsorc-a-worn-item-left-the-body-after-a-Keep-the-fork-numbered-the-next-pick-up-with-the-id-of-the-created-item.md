---
id: TASK-810
title: 'recallsorc: a worn item left the body after a Keep: the fork numbered the next pick-up with the id of the created item'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-22T18:37:00Z'
labels:
  - d2d
  - bug
dependencies: []
priority: medium
ordinal: 371000
card: "recallsorc"
owns:
  - "internal/d2dsrc/items.go"
  - "internal/d2dsrc/source.go"
  - "internal/d2dsrc/dupid_test.go"
  - "internal/bot/lifelog.go"
  - "internal/bot/observe.go"
  - "internal/loot/pickup.go"
  - "docs/vault.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Found by the equipswap run of 2026-11-22. Done when the four boxes are ticked.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Root cause named from the live log (manage.recall-now Sorceress 3: the worn cap and the next picked potion both carried unit id 4)
- [ ] #2 Fork fix and a test that fails without it (vault_verbs_d2d.py, branch didle/recallsorc-20261122)
- [ ] #3 Bot reader names carried items that share a unit id (must-fail test, Paladin and Sorceress)
- [ ] #4 manage.recall-now green on the lab d2d with the fix, both classes (live, CHEATED setups)
<!-- AC:END -->
