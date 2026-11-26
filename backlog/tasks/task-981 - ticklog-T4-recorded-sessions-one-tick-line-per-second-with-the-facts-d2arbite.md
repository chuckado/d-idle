---
id: TASK-981
title: 'ticklog: T4 recorded sessions: one tick line per second with the facts, d2arbite'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:56:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: high
ordinal: 469000
card: "ticklog"
owns:
  - "docs/design/ticklog-t4.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Cause class killed: A before live; gives the arbiter migration an outcome measure. Roadmap row T4 (docs/research/fable-logic-review-2026-11-25.md 5.3, 7.3). One `tick` line per second with the facts the shadow adapter already collects (internal/bot/shadow.go), rate-limited, off by default; cmd/d2arbiter -follow replays them; makes the arbiter fixtures real (the Dark Wood fixture's facts are assumed, docs/arbiter.md). Risk: low (log volume). Enable criteria: tick log on both sandboxes for a night, replay runs without a panic.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Tick line behind a knob with a ledger row
- [ ] #2 Replay of one real sandbox night with -follow documented
<!-- AC:END -->
