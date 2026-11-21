---
id: TASK-776
title: 'cainscroll: Cain scroll step follows the quest record bits, quest items never sold'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T14:38:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 359000
card: "cainscroll"
owns:
  - "docs/probes/cain-tree-scroll-2026-11-21.md"
  - "tools/provingground/scenarios/cain.json"
  - "internal/task/act1.go"
  - "internal/task/world.go"
  - "internal/task/cainword_test.go"
  - "cmd/d2bot/task.go"
  - "internal/loot/questcodes.go"
  - "internal/loot/questcodes_test.go"
  - "internal/loot/rulesadapt.go"
  - "internal/town/stashfull.go"
  - "internal/town/tidy.go"
  - "internal/town/questnever_test.go"
  - "data/questitems"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Sandbox hero Doofer parked search_for_cain/scroll: tree clicked, no scroll (record 0x0018). Done when the scenario is green and merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Diagnosis doc: where the scroll is, root cause, bit table
- [ ] #2 Scenario cain.scroll-already-taken RED on the unfixed bot, GREEN with the fix; cain.scroll and cain.chain stay green
- [ ] #3 Board evidence follows record bit 4; quest items of Acts 1 and 2 never sold; must-fail tests on two builds
- [ ] #4 -race -count=20, check.sh, contracts, backlog, leakcheck clean
<!-- AC:END -->
