---
id: TASK-910
title: 'toolstrade37: why the Tower Cellars are not farmed: the zone model, and a farm picker log line'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-24T00:12:00Z'
labels:
  - bot
  - bot-fix
dependencies: []
priority: medium
ordinal: 415000
card: "toolstrade37"
owns:
  - "internal/bot/goal.go"
  - "internal/bot/plan.go"
  - "internal/bot/plan_test.go"
  - "docs/probes/tools-trade-2026-11-22.md"
exclusive: []
contracts: []
result: "Cause found from the live zone choice line (survival margin model rejects the cellars); a farm picker log line added; no default changed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: S. Done when merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The picker logs the exclusions and the plan rows reached; the cause of the missing Cellars is recorded with the live zone choice numbers
<!-- AC:END -->
