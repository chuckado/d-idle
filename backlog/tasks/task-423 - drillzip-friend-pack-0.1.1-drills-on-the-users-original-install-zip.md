---
id: TASK-423
title: 'drillzip: friend pack 0.1.1 drills on the user''s original install zip'
status: Done
assignee:
  - credit-runner
created_date: '2026-11-14T15:40:00Z'
updated_date: '2026-11-14T15:55:00Z'
labels:
  - container
  - research
dependencies: []
priority: medium
ordinal: 255000
card: "drillzip"
owns:
  - "docs/probes/drill-zip-2026-11-14.md"
exclusive: []
result: "Pack 0.1.1 drilled on the user install zip (verified live, worker host): D1, D2, report, stop/start, 7 failure paths pass; D3/D4 not run (no newer signed release). Bug: an unquoted game path with a space (Diablo II) breaks every ./didle command; pack works without the video MPQs."
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the user asked for the original install as the drill target. Done when: the record is in docs/probes/drill-zip-2026-11-14.md. Consumer: friends pack 0.1.2 fixes.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Drills D1, D2 and the failure paths run on the zip target and recorded
- [x] #2 Doc docs/probes/drill-zip-2026-11-14.md written, leakcheck clean
<!-- AC:END -->
