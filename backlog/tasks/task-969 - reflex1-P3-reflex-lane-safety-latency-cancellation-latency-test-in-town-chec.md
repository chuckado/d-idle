---
id: TASK-969
title: 'reflex1: P3 reflex lane: safety latency, cancellation-latency test, in-town chec'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:53:00Z'
labels:
  - bot
  - feature
dependencies:
  - TASK-967
priority: high
ordinal: 458000
card: "reflex1"
owns:
  - "docs/design/reflex-p3.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Cause class killed: the bug:true safety deaths (3.1 rows 3 and 4 of the review). Roadmap row P3 (docs/research/fable-logic-review-2026-11-25.md 7.1, 7.2 PR 3). (a) measure safety latency in the autopsy and minute line; (b) T6 cancellation-latency test: every action registered with do returns within a bound after cancel; (c) a drink sent by the safety check itself under the input lock without cancelling the action, with a hard deadline after which portal and chicken act anyway; (d) town chores verify in-town before each NPC step. Changes the rule 'the guard only watches' in docs/bot.md for one action (listed as a proposed change in the review 7.5; owner word needed before step c). Scenario from the 02:51 death. Risk: medium. Order: a, b, d first (no rule change); c after the owner's yes. Enable criteria: safety latency under 2 s on every minute line for 3 sandbox days; no bug:true death in 7 days.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Latency measured in autopsy and minute line (a); cancellation test (b); in-town check (d)
- [ ] #2 Step (c) only after the owner's word on the docs/bot.md rule; scenario from the 02:51 death
<!-- AC:END -->
