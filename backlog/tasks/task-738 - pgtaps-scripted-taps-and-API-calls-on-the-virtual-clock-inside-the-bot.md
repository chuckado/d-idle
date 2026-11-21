---
id: TASK-738
title: 'pgtaps: scripted taps and API calls on the virtual clock inside the bot'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-21T04:46:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 344000
card: "pgtaps"
owns:
  - "cmd/d2bot/ladder.go"
  - "cmd/d2scenario/tap.go"
  - "internal/scenario/taps.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: L

Design section 2.8 step 3, 7 card 4. Medium risk: touches the bot answer path. Design: docs/design/fable-review-2026-11-21/proving-ground-gate.md. Owner approved 2026-11-21 (all 10 questions as recommended).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 --scripted-taps file with rules evaluated on the stepping goroutine
- [ ] #2 HTTP tapper stays for sandbox e2e, those runs are never stable
- [ ] #3 two-run pair of one board scenario has equal D1 and equal Start tick
<!-- AC:END -->
