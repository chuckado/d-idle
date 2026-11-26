---
id: TASK-967
title: 'invmon: P1 invariant monitor: about twelve checks per tick, log-only'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-26T03:52:00Z'
labels:
  - bot
  - feature
dependencies: []
priority: high
ordinal: 456000
card: "invmon"
owns:
  - "docs/design/invmon-p1.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: S

Cause classes killed: K (not visible), finds A, B, C, E live and in lab. Roadmap row P1 (docs/roadmap-fable.md; docs/research/fable-logic-review-2026-11-25.md 7.1). Table of about twelve checks run once per tick on the snapshot and loop state; a hit logs one line `invariant` with name and facts and counts in the minute line. Log only, no behaviour. First rows: resupplying outside town over 10 s; field entry under the resume line; safety verdict not carried out within 2 s; same step refused 10 times with no fact changed; trips per hour over N; gold under kit cost for 2 h; life zero while alive. Test with two builds (Sorceress and Paladin facts). Risk: none. Order: first; P5 flip and P6 outcome rule need its counts. Enable criteria: no knob needed (log only); counts read on both sandboxes for 3 days. Done when: monitor merged, counts visible in the minute line, check list in the doc.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Monitor table and log line with a test per check (two builds)
- [ ] #2 Minute-line counters; design doc lists each check and its threshold evidence
<!-- AC:END -->
