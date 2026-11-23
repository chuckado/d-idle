---
id: TASK-842
title: 'toolstrade4: pocket and approach cells keep off other levels'' border strips'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T03:08:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 383000
card: "toolstrade4"
owns:
  - "internal/nav/exit.go"
  - "internal/nav/gatedoor.go"
  - "internal/nav/gatedoor_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Monastery Gate re-entry no longer slips into the Tamoe Highland (lab scenario act1.tools.malus Paladin 11, sandbox knobs: Entered Outer Cloister on the re-entry); chunk not yet green: the hero then goes home from the Outer Cloister (route guard, portal refused by the crowd rule)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. The Monastery Gate arrival pocket sent the hero onto the south strip and into the Tamoe Highland. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 pocketTarget and nearestReachableTo skip cells on or beside the strips toward other levels; must-fail test, two classes
<!-- AC:END -->
