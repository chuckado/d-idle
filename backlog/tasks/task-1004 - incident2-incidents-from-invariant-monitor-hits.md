---
id: TASK-1004
title: 'incident2: incidents from invariant monitor hits'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T17:49:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: medium
ordinal: 482000
card: "incident2"
lane: next
owns:
  - "internal/bot/incident.go"
  - "internal/bot/incident_test.go"
  - "internal/bot/invariants.go"
  - "internal/scenario/incident.go"
  - "internal/scenario/incident_test.go"
  - "cmd/d2scenario/incident.go"
  - "cmd/d2scenario/incident_test.go"
  - "docs/probes/incident-scenarios.md"
exclusive: []
contracts: []
result: "An invariant hit (hard or soft) writes a rate limited, SETUP tagged pre-state file next to the death incidents; d2scenario incident replays it (or the whole directory) as a report-only nightly row that is red when the same invariant fires again. Unit tests only, no lockstep run, not verified live."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

P8 remainder (Fable review 2026-11-25).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 monitor hit writes a rate limited SETUP pre-state file
- [x] #2 d2scenario incident replays monitor files as report-only rows
<!-- AC:END -->
