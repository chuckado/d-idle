---
id: TASK-573
title: 'videofeed4: stream monitor overlay, stable presentation clock, lighter d2dview feed (cgroup-throttle hitching)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T01:58:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 299000
card: "videofeed4"
owns:
  - "internal/watch/palfeed.go"
  - "internal/watch/palfeed_test.go"
  - "internal/watch/views.go"
  - "internal/watch/testdata/panel_harness.js"
  - "internal/watch/testdata/video_pal_panel.js"
  - "internal/watch/web/palfeed.js"
  - "internal/watch/web/streamstats.js"
  - "internal/watch/web/panels/video"
  - "internal/watch/palfeed"
  - "cmd/d2dview"
  - "tools/videofeed"
  - "docs/video-feed.md"
  - "internal/watch/testdata/video_pal_panel.js"
  - "internal/watch/testdata/video_panel.js"
  - "internal/watch/http_view.go"
  - "tools/uishots"
  - "tools/uishots.sh"
exclusive: []
contracts: []
result: "Measured cause (read-only samples in the sandbox container): the 1-core CPU quota is saturated by the bot's own bursts (10-47 s at 64-71 percent of the core, 67-87 s apart), the container is throttled in 7-10 of 10 periods, d2d starves from 33 to 18-22 ticks/s in 2-7 s runs every 3-16 s. Added: stream monitor overlay (30 s charts, time axis, JSON copy, both players), host numbers in /api/palfeed, SlowGate hysteresis for the low-fps hint, a stable deeper presentation clock (66-200 ms) with a synthetic stall regression test, rate halved while throttled, pooled hub frames. Verified offline and in headless Chromium; not on the sandbox."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 measured cause of the hitching documented (container cgroup throttle, bot CPU bursts) with sandbox read-only samples
- [x] #2 on-page stream monitor (30 s rolling charts, JSON copy, both players), off by default
- [x] #3 presentation clock stable (no oscillation) and deep enough to hide 100 ms stalls; regression test with a synthetic stall trace
- [x] #4 d2dview feed CPU reduced (no per-frame allocation) and server stats in the capability reply
<!-- AC:END -->
