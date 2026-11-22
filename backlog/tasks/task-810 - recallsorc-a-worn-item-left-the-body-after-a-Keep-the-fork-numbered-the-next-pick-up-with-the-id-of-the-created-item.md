---
id: TASK-810
title: 'recallsorc: a worn cap left the body on a pickup: unwear matched the ground item id against a worn item with the same number'
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
  - "internal/loot/loot_test.go"
  - "docs/vault.md"
exclusive: []
contracts: []
result: "Root cause (V live, bot log): loot.unwear looked the picked Strangling Gas Potion up by unit id, found the worn cap with the same id (4), and took the cap off because a throwing weapon is no weapon for the class; the Sorceress seeds hit the id match, the Paladin seeds did not. Fixed (unwear acts only on a unit of the same base that was not worn before the pickup), must-fail test on Sorceress and Paladin (needs game tables, run on the worker host). manage.recall-now GREEN live on the lab d2d devp19, Sorceress 3, 11, 5 and Paladin 3, 11, 5 (CHEATED setups, mechanics only). Second defect found on the way and fixed in the private fork (branch didle/recallsorc-20261122, build devp21, not deployed, sandbox untouched): item create numbers its item max(carried)+1 without advancing World::next_item_id, so the next pick-up got the same unit id; test fails on devp19, passes on devp21. The bot names shared carried ids in the log once per id and logs every change of the worn set."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Found by the equipswap run of 2026-11-22. Done when the four boxes are ticked.

Amended 2026-11-22: the cap loss was a bot defect (unwear by id), found with the worn-set log; the fork id defect is a second, separate fix.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Root cause named from the live log (manage.recall-now Sorceress 3: unwear matched the picked potion's ground id 4 against the worn cap, also id 4)
- [x] #5 Must-fail test on Sorceress and Paladin: a pickup whose id equals a worn item's id never takes it off (internal/loot)
- [x] #2 Fork fix for the duplicate unit id after item create and a test that fails without it (vault_verbs_d2d.py, branch didle/recallsorc-20261122; build devp21, not deployed)
- [x] #3 Bot reader names carried items that share a unit id (must-fail test, Paladin and Sorceress)
- [x] #4 manage.recall-now green on the lab d2d devp19 with the fix, both classes, 3 seeds each (live, CHEATED setups)
<!-- AC:END -->
