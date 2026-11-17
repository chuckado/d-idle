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
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 measured cause of the hitching documented (container cgroup throttle, bot CPU bursts) with sandbox read-only samples
- [ ] #2 on-page stream monitor (30 s rolling charts, JSON copy, both players), off by default
- [ ] #3 presentation clock stable (no oscillation) and deep enough to hide 100 ms stalls; regression test with a synthetic stall trace
- [ ] #4 d2dview feed CPU reduced (no per-frame allocation) and server stats in the capability reply
<!-- AC:END -->
