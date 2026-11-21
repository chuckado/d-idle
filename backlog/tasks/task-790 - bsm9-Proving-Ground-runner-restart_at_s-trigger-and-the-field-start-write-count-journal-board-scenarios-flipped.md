---
id: TASK-790
title: 'bsm9: Proving Ground runner restart_at_s trigger and the field-start write count (journal), board scenarios flipped'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T19:46:00Z'
labels:
  - bot
  - chore
dependencies: []
priority: medium
ordinal: 362000
card: "bsm9"
owns:
  - "cmd/d2scenario/run.go"
  - "cmd/d2scenario/restart.go"
  - "cmd/d2scenario/restart_test.go"
  - "cmd/d2scenario/fieldledger.go"
  - "cmd/d2scenario/fieldledger_test.go"
  - "internal/scenario/scenario.go"
  - "tools/provingground/scenarios/board.json"
  - "docs/probes/boardsm-progress-2026-11-21.md"
exclusive: []
contracts: []
result: "Runner: trigger.restart_at_s stops the bot (SIGTERM, save and quit) and restarts it on the same hero and plan file with the rest of the virtual budget; a field start now reports writes after start from the bot d2dctl journal (accepted debug write verbs only). Unit tested. Live (second worker, CHEATED setups): restart-mid-chain, handin.without-run and handin.field-tap GREEN on Paladin 11 (the lab word 0x0002 read as needed because Interpret wants a started bit; 0x000e fixes the setup); hold.restock and restock-stash-gold GREEN (bsm7); hold.broke asserts the recovery (no park); hold.too-weak, hold.timeout, park.persist-restart, regress-proxy stay pending with their reasons (setups that cannot be made, U). Sorceress runs of these were not done: lab pace."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 trigger.restart_at_s stops the bot (SIGTERM, save and quit) and starts it again on the same hero and state dir with the rest of the budget
- [x] #2 a field start counts the bot's accepted debug verbs from its d2dctl journal as writes after start
- [x] #3 handin.*, restart-mid-chain and hold.restock scenarios run live (Paladin 11); the rest stay pending with reasons
<!-- AC:END -->
