---
id: TASK-868
title: 'toolstrade10: GoTo opens doors in a row; tools scenario hero wears a belt'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T07:31:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 391000
card: "toolstrade10"
owns:
  - "internal/nav/gatedoor.go"
  - "internal/nav/gatedoor_test.go"
  - "tools/provingground/scenarios/act1more.json"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "GoTo walks a corridor of doors one door at a time; unit-tested; lab Smith kill still pending."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Lab: the Smith's room lies behind a corridor of doors; one door opened, then closest reachable point again. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 GoTo opens the near door when several doors stand between the hero and the spot; must-fail test, two classes
<!-- AC:END -->
