---
id: TASK-858
title: >-
  toolstrade7: boss not found says what stood in view; tools scenario starts
  started
status: Done
assignee:
  - credit-runner
created_date: '2026-11-23T05:32:00Z'
updated_date: '2026-11-23T05:56:00Z'
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
result: "Smith anchor resolves by name (quests.layoutRead + debug levelmonsters) and a not-found boss run notes the monsters in view; unit-tested, lab run with the knob on pending."
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
