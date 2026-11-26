---
id: TASK-962
title: 'reflex1a: safety latency measure and cancellation-latency test'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:42:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 455000
card: "reflex1a"
owns:
  - "internal/bot/safetylat.go"
  - "internal/bot/safetylat_test.go"
  - "internal/bot/autopsy.go"
  - "internal/bot/loop.go"
  - "internal/bot/observe.go"
  - "internal/combat/cancellat_test.go"
  - "docs/safety-latency.md"
exclusive: []
result: "Safety latency (verdict to input) is on the minute line (safetyN, safetyLatAvgMs, safetyLatMaxMs, safetySlow) and in the death autopsy (safetyLatMs); T6 cancel test covers combat Clear and Kill on two classes; nav, town and loot actions listed as not covered in docs/safety-latency.md. Offline only, not verified live."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Fable review P3 (a)(b). Done when measured on two classes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 latency of each safety verdict to the input sent is in the autopsy and the minute line
- [x] #2 T6: every listed action returns within N ms of a cancel (fakes); actions that do not are listed
<!-- AC:END -->
