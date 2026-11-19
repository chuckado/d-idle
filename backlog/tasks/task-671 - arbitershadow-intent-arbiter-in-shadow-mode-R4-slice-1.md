---
id: TASK-671
title: 'arbitershadow: intent arbiter in shadow mode (R4 slice 1)'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-19T17:22:00Z'
updated_date: '2026-11-19T18:11:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 328000
card: "arbitershadow"
owns:
  - "internal/arbiter"
  - "cmd/d2arbiter"
  - "internal/bot/shadow.go"
  - "internal/bot/shadow_test.go"
  - "docs/arbiter.md"
  - "internal/env/tuning_bot.go"
  - "docs/knobs.md"
  - "internal/bot/loop.go"
  - "internal/bot/runner.go"
  - "internal/bot/observe.go"
exclusive: []
contracts: []
result: "Shadow arbiter built and tested offline (replay of the 11-18 Dark Wood cycle: chain 412 area changes, arbiter 2). NOT verified live: the sandbox shadow log is the next read. Build-specific data (belt line, flee line) comes as facts from the Tune tables."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Roadmap R4 slice 1. Shadow only, chain unchanged. Done when merged and the sandbox can enable bot.arbiterShadow.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 internal/arbiter: intents, lease, priced travel, town batching, trace; scenario tests incl. the prototype's
- [x] #2 bot.arbiterShadow (default 0) runs the arbiter beside runGoal, logs divergences, no chain change
- [x] #3 replay tool and ping-pong fixture test
- [x] #4 docs/arbiter.md with migration plan
<!-- AC:END -->
