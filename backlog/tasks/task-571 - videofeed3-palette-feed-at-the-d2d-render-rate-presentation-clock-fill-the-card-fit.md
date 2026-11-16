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
  - "tools/uishots"
  - "tools/uishots.sh"
  - "internal/watch/web/panels/video/panel.css"
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
result: "Cause of 14-15 fps: d2dview's ring divisor (a whole k so the JPEG back-off fits, -fps 15 gives every 4th draw, -fps 25 every 3rd = 20 fps) applied to the lossless feed too; now 33 ms (30 fps) while a canvas viewer is connected, JPEG keeps its own cadence; hub pace is a schedule. Jitter: presentation clock (game-time stamps, adaptive 1-2 frame delay, latest wins, stats), measured in headless Chromium on the builder: draw interval sd 17.3 -> 10.5 ms (floor 8.2 at 25 Hz on 60 Hz), merged frames 18 -> 1. Fit: canvas fills the card at any scale (fitFill, dpr aware), smooth unless Sharp pixels or integer scale. Verified offline and in headless Chromium, not on the sandbox; Firefox not measurable on the builder."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 feed rate follows the d2d render rate (25 to 30 fps) when the viewer asks; cause of the 14-15 fps cap found and fixed
- [x] #2 presentation clock: stamped frames, adaptive 1-2 frame jitter buffer, latest wins, presented vs delivered vs dropped stats, no per-frame allocation
- [x] #3 canvas and MJPEG fill the card (contain, fractional scale, dpr aware), Sharp pixels/Smooth toggle, fullscreen works
- [x] #4 node tests for pacing; measured delivered fps and presentation jitter in headless Chromium
<!-- AC:END -->
