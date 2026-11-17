---
id: TASK-581
title: 'palfeedflaky: make the palfeed guard tests deterministic'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T03:19:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 301000
card: "palfeedflaky"
owns:
  - "internal/watch/palfeed"
exclusive: []
contracts: []
result: "TestGuardPacesBigMessagesAndCountsClients asserted real elapsed time and failed under CI load. The guard arithmetic is now a pure function (pace) tested with fixed times; the stream tests assert counters and ordering only (Options.Heartbeat makes the heartbeat test instant, lock-step delivery replaces the 30 Hz timing count). Verified on the builder with -race -count=300 -cpu 1,2,4 and the whole package under CPU contention."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 guard pacing arithmetic is a pure function tested with fixed times; no wall-clock assertions in palfeed tests
- [x] #2 go test -race -count=300 -cpu 1,2,4 passes on the builder
<!-- AC:END -->
