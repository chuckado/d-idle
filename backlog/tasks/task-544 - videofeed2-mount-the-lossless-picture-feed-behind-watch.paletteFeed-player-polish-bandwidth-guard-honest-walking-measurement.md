---
id: TASK-544
title: 'videofeed2: mount the lossless picture feed behind watch.paletteFeed, player polish, bandwidth guard, honest walking measurement'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-16T05:24:00Z'
labels:
  - watch
  - feature
dependencies: []
priority: medium
ordinal: 287000
card: "videofeed2"
owns:
  - "internal/watch/palfeed"
  - "internal/watch/palfeed.go"
  - "internal/watch/web/palfeed.js"
  - "internal/watch/web/panels/video"
  - "internal/env/tuning_watch.go"
  - "cmd/d2dview"
  - "tools/videofeed"
  - "docs/video-feed.md"
  - "docs/probes/video-feed-2026-11-15.md"
  - "internal/env/tuning.go"
  - "internal/watch/serve.go"
  - "internal/watch/hub.go"
  - "internal/watch/views.go"
  - "internal/watch/glance.go"
  - "internal/watch/video.go"
  - "internal/watch/http_view.go"
  - "internal/watch/palfeed_test.go"
  - "internal/watch/watchlayout_test.go"
  - "internal/watch/testdata/video_pal_panel.js"
  - "internal/watch/testdata/panel_harness.js"
  - "cmd/d2bot/ladder.go"
  - "tools/knobdump/main.go"
  - "docs/knobs.md"
  - "docs/friend-guide.md"
  - "tools/friendpack/compose.yml"
exclusive: []
contracts: []
result: "Mounted behind watch.paletteFeed / D2_PALETTE_FEED=1 (default off) in d2dview and the d2bot tap relay; canvas player with integer scaling, MJPEG fallback, Low rate; per-client bandwidth guard, heartbeat. Measured on a 2860-frame walking-heavy lab recording (builder, lab hero): lossless feed 0.57 MB/s against 1.79 MB/s JPEG q72 at half the time walking (3.1x less), town walk only 1.4x less; the per-frame JPEG hybrid is worse and not built. Verified offline (Go -race, node), NOT live: the sandbox A/B is the checklist in docs/video-feed.md section 10; the friends image lacks framefeed."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: user go 2026-11-15 to finish the lossless picture feed. Done when: the feed is mounted behind the knob (default off), the page falls back to MJPEG, the guard and tests are in, and the walking-heavy measurement states the honest verdict (docs/probes/video-feed-2026-11-15.md). Not deployed; the sandbox A/B is docs/video-feed.md section 10.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 knob watch.paletteFeed registered (default 0), docs/knobs.md regenerated; feed mounted in d2dview and the watch relay behind it
- [x] #2 page picks the canvas player from glance stream.formats and falls back to MJPEG on any trouble; Low rate and fps stat apply to both
- [x] #3 bandwidth guard, key-size pacing, write deadline, heartbeat; Go tests with -race on palfeed, node tests
- [x] #4 walking-heavy measurement and honest verdict in the probe doc
- [x] #5 docs: friend guide switch, friends image lacks framefeed, live A/B checklist (not run)
<!-- AC:END -->
