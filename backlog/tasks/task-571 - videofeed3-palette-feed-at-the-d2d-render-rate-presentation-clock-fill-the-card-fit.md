---
id: TASK-571
title: 'videofeed3: palette feed at the d2d render rate, presentation clock, fill-the-card fit'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T23:08:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 297000
card: "videofeed3"
owns:
  - "internal/watch/web/palfeed.js"
  - "internal/watch/web/panels/video"
  - "internal/watch/palfeed"
  - "cmd/d2dview"
  - "tools/videofeed"
  - "docs/video-feed.md"
  - "internal/watch/testdata/video_pal_panel.js"
  - "internal/watch/testdata/video_panel.js"
  - "internal/watch/testdata/panel_harness.js"
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
- [ ] #1 feed rate follows the d2d render rate (25 to 30 fps) when the viewer asks; cause of the 14-15 fps cap found and fixed
- [ ] #2 presentation clock: stamped frames, adaptive 1-2 frame jitter buffer, latest wins, presented vs delivered vs dropped stats, no per-frame allocation
- [ ] #3 canvas and MJPEG fill the card (contain, fractional scale, dpr aware), Sharp pixels/Smooth toggle, fullscreen works
- [ ] #4 node tests for pacing; measured delivered fps and presentation jitter in headless Chromium
<!-- AC:END -->
