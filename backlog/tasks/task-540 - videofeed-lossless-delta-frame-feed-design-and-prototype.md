---
id: TASK-540
title: 'videofeed: lossless delta frame feed (design and prototype)'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T03:51:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 286000
card: "videofeed"
owns:
  - "internal/watch/palfeed"
  - "internal/watch/web/palfeed.js"
  - "docs/video-feed.md"
  - "docs/probes/video-feed-2026-11-15.md"
  - "tools/videofeed"
exclusive: []
contracts: []
result: "Design docs/video-feed.md + prototype internal/watch/palfeed and web/palfeed.js (not wired, knob watch.paletteFeed reserved). V on 630 real d2d frames (lab d2d on the worker host): lossless, p50 9.7 KB, p95 55 KB per frame at 25 fps (JPEG q72 69/103 KB), 0.49 MB/s mean; encode 0.5 ms p50, browser decode 1.2 ms p50. d2d frame is RGBA (not palettised, 340 colours/frame): no d2d patch needed, the existing framefeed ring is the source. Not verified live in the sandbox, not wired."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: the viewer's game video (MJPEG, 1.7 to 1.9 MB/s, lossy, 10 ms JPEG per frame on one core) can be exact and far cheaper if the d2d frame ring is sent as tile deltas. Done when: design doc, additive prototype (flag-off, nothing wired), tests and numbers from real recorded d2d frames exist.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 design doc docs/video-feed.md
- [x] #2 encoder/decoder + tests, JS player prototype + node tests
- [x] #3 numbers from real recorded d2d frames
<!-- AC:END -->
