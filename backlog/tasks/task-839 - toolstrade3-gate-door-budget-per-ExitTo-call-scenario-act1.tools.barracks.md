---
id: TASK-839
title: 'toolstrade3: gate door budget per ExitTo call; scenario act1.tools.barracks'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T02:39:00Z'
updated_date: '2026-11-23T03:01:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 382000
card: "toolstrade3"
owns:
  - "internal/nav/exit.go"
  - "internal/nav/nav.go"
  - "internal/nav/gatedoor_test.go"
  - "tools/provingground/scenarios/act1more.json"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Door budgets reset per ExitTo call (the Monastery Gate door was refused from the third visit on); unit-tested, lab scenario runs recorded in docs/probes/tools-trade-2026-11-22.md."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Follow-up of toolstrade2: the third visit to the Monastery Gate refused the door (cut off, no way out). Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The same gate door opens on every visit (budget per ExitTo call); must-fail test, two classes
- [x] #2 Scenario act1.tools.barracks exists (field start in the Outer Cloister)
<!-- AC:END -->
