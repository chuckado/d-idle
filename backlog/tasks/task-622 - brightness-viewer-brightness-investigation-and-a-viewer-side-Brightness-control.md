---
id: TASK-622
title: 'brightness: viewer brightness investigation and a viewer-side Brightness control'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T15:31:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 310000
card: "brightness"
owns:
  - "docs/probes/viewer-brightness-2026-11-17.md"
  - "internal/watch/web/panels/video/brightness.js"
  - "internal/watch/brightness_test.go"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Size: M. User: viewer overly bright. Done when the cause is measured and the control ships.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 probe doc with numbers: which component is brighter
- [ ] #2 viewer Brightness control (Dark/Default/Bright), per browser, default 100%
- [ ] #3 regression test: viewer canvas path keeps reference luminance
- [ ] #4 gates, mergeq
<!-- AC:END -->
