---
id: TASK-903
title: 'malus: quest drop retries and ground list, tools scenario stand-in, Malus finding'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-23T22:14:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 409000
card: "malus"
owns:
  - "internal/boss/run.go"
  - "internal/boss/wire_test.go"
  - "cmd/d2bot/boss.go"
  - "tools/provingground/scenarios/necro.json"
  - "docs/probes/necromancer-act1-drill-2026-11-22.md"
exclusive: []
contracts: []
result: "boss loot phase: 3 retries of 3 s for a quest item (was one of 2 s), ground list logged when it is missing; necro.act1-tools is a clvl 40 stand-in (life floor 300), reach:50. Finding: after the army kills the Smith no item lies within 25 subtiles for 9 s (run 165530-314a): cause open (no own killing blow, Charsi state, or lab pets); Paladin control inconclusive (life floor 400). Stage 2b RED. Mana-kit park is a consequence of the reset loop."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Quest drop waits three retries, unit test fails with the old single retry
- [x] #2 Ground list logged when a quest item is missing
- [x] #3 Stage 2b verdict and the Malus finding recorded (still RED)
<!-- AC:END -->
