---
id: TASK-830
title: 'toolstrade: Tools of the Trade stuck at the Monastery Gate border exit'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T00:52:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 379000
card: "toolstrade"
owns:
  - "internal/nav/exit.go"
  - "internal/nav/gatedoor.go"
  - "internal/nav/gatedoor_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "ExitTo opens the gate door after a walk that ends short (Monastery Gate to Outer Cloister, first attempt now enters); verified in a lab scenario run (CHEATED setup), full chunk NOT green: arrival fight and a town re-entry are separate findings."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. User report: tools_of_the_trade/to-barracks parks gave_up. Done when the fix is merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Root cause found from a scenario run (act1.tools.malus), recorded in docs/probes/tools-trade-2026-11-22.md
- [x] #2 ExitTo looks for the gate door when the walk ends at the closest reachable point; must-fail test, two classes
- [x] #3 Scenario act1.tools.malus verdicts with the sandbox knobs and with both off recorded
<!-- AC:END -->
