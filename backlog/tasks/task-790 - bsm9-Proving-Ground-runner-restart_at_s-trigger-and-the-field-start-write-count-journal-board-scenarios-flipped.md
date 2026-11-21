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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 trigger.restart_at_s stops the bot (SIGTERM, save and quit) and starts it again on the same hero and state dir with the rest of the budget
- [ ] #2 a field start counts the bot's accepted debug verbs from its d2dctl journal as writes after start
- [ ] #3 handin.* , hold.* and restart-mid-chain scenarios run live
<!-- AC:END -->
