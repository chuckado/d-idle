---
id: TASK-968
title: 'outcome1: P2 typed step outcome (Done, Progress, Refused, Failed, Interrupted),'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:52:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: high
ordinal: 457000
card: "outcome1"
owns:
  - "docs/design/outcome-p2.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Cause classes killed: B whole, much of A and K; ground for P4, P6, P7, P9. Roadmap row P2 (docs/research/fable-logic-review-2026-11-25.md 7.1, 7.2 PR 2). A step ends as Done, Progress, Refused{need, wake}, Failed{fault} or Interrupted{cause}; one handler: a refusal never increments a failure counter, is logged once per change as INFO `step refused` with need= and wake=, and becomes the hero's next concern (trip, farm, level target) or a park with a wake condition. First slice: boardSlice and taskStep return it. Deletes over time: planningFault word match, errAreaGated special cases, watch hub wording match, most `*Until` backoffs. Contract: docs/contracts/bot-loop.md changes in the implementing card. Scenario: the early-road refusal of 2026-11-26 yields one line and one kit trip, not 226 warnings. Risk: low to medium (touches do, boardSlice, taskStep). Order: after invmon starts, before P4, P6, P7, P9, P12. Enable criteria: refusal-storm monitor row (P1) at 0 for 3 sandbox days; the knob (default 0 then 1) is the implementing card's call.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Outcome type and one handler behind a knob, default off, with a ledger row
- [ ] #2 Scenario from the 2026-11-26 early-road refusal: one INFO line, no failure counter
<!-- AC:END -->
