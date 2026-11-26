---
id: TASK-972
title: 'townflip: P5 finish and flip town.plan: stage 2 on sandboxes, then stages 3 to 5'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:54:00Z'
labels:
  - bot
  - chore
dependencies:
  - TASK-955
priority: high
ordinal: 460000
card: "townflip"
owns:
  - "docs/design/townflip-p5.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Cause class killed: C supply and economy (49, the largest). Roadmap row P5 (docs/research/fable-logic-review-2026-11-25.md 7.1; ADR 0028 as decided). `town.plan` is built to stage 3 and default 0 (town1 to town5 merged or in flight); this card owns the FLIP, not the stage code: set stage 2 on both sandboxes (quest and farm), scenario set green, then flip stage 2 with a 24 h invariant-monitor watch and auto-revert on a hard breach, then flip the default to 2, then stages 3, 4 (town6) and 5 (town8) one at a time, then town11 deletes the old counters and resource.* knobs. Updates the dark-launch row on each step. Depends: P1 counts exist before the flip; stage code cards town5, town6, town7, town8. Risk: medium, mitigated by before/after counts. Enable criteria (from data/darklaunch.json): stage 2 scenario set green, scenario/drill green, flip on the sandboxes, 24 h invariant-monitor watch, auto-revert on a hard breach (the standard rule of user decision 2026-11-26); later stages the same, one at a time. Owner pause rule of ADR 0028: user looks before purse, sell, route stages.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Stage 2 set on both sandboxes with the 24 h watch recorded
- [ ] #2 Default flipped per stage with counts in the card result; darklaunch row updated each step
<!-- AC:END -->
