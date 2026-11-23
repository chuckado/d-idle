---
id: TASK-858
title: 'toolstrade7: boss not found says what stood in view; tools scenario starts started'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T05:32:00Z'
labels:
  - bot
  - probe
dependencies: []
priority: medium
ordinal: 387000
card: "toolstrade7"
owns:
  - "internal/boss/run.go"
  - "tools/provingground/scenarios/act1more.json"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Diagnostic only: the next lab run names the monsters in the Barracks; the Smith is still not found in the lab (V runs, level 28 hero, knobs on and off)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Lab run act1.tools.malus reaches the Barracks and the Smith is not found (preset 754 absent): the note says what is there. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 A boss run that ends in boss-not-found notes the class, unique number and name of the monsters in view
<!-- AC:END -->
