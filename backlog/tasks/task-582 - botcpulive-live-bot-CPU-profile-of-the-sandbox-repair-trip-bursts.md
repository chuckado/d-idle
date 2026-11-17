---
id: TASK-582
title: 'botcpulive: live bot CPU profile of the sandbox (repair-trip bursts)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T03:22:00Z'
labels:
  - ops
  - probe
dependencies: []
priority: medium
ordinal: 301000
card: "botcpulive"
owns:
  - "docs/probes/bot-cpu-live-2026-11-16.md"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Sandbox profiling per user go 2026-11-16. Done when the doc is merged.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Live baseline CPU profile captured and analysed
- [ ] #2 Burst-window profile captured (NOT met: one-shot profiler)
- [ ] #3 Sandbox restored (profiling off, cpus decided)
<!-- AC:END -->
