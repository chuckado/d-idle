---
id: TASK-776
title: >-
  cainscroll: Cain scroll step follows the quest record bits, quest items never
  sold
status: Done
assignee:
  - credit-runner
created_date: '2026-11-21T14:38:00Z'
updated_date: '2026-11-21T20:31:00Z'
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
result: "Sandbox hero (record 0x0018) stalled at the tree because the board steps read only the pack, not the quest word; bit 4 (stones done) means the scroll was used and the tree never drops again. Fixed: World carries the record word, Cain field steps follow bit 4; Acts 1 and 2 quest items (data/questitems) are never sold or dropped. Second worker Proving Ground: cain.scroll-already-taken RED unfixed, GREEN fixed (Paladin 11, Sorceress 3, key variant); cain.scroll Paladin 11 and cain.chain Paladin 11 GREEN; cain.scroll Sorceress 3 RED on both trees (pre-existing). Not verified on the sandbox. Doc docs/probes/cain-tree-scroll-2026-11-21.md."
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
- [x] #1 Diagnosis doc: where the scroll is, root cause, bit table
- [x] #2 Scenario cain.scroll-already-taken RED on the unfixed bot, GREEN with the fix; cain.scroll and cain.chain stay green
- [x] #3 Board evidence follows record bit 4; quest items of Acts 1 and 2 never sold; must-fail tests on two builds
- [x] #4 -race -count=20, check.sh, contracts, backlog, leakcheck clean
<!-- AC:END -->
