---
id: TASK-590
title: >-
  videofeed5: presentation stall (animation-frame clock, withheld frames, stuck
  inflate) and the build id stamp
status: Done
assignee:
  - credit-runner
created_date: '2026-11-17T04:56:00Z'
updated_date: '2026-11-17T06:13:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 303000
card: "videofeed5"
owns:
  - "internal/watch/web/palfeed.js"
  - "tools/videofeed"
  - "docs/video-feed.md"
  - "tools/dev-update.sh"
  - "tools/dev-update-host.sh"
  - "deploy/d2d-container/context.sh"
  - "internal/buildinfo"
exclusive: []
contracts: []
result: "The Player compared due times with the animation frame's timestamp while building them from performance.now: a skewed or withheld animation-frame clock held every frame back then released the backlog in one draw (frames arrive, nothing shown, dropped spikes). Now one clock (performance.now), a 100 ms fallback timer while animation frames are withheld, a stuck-inflate watchdog that reconnects, events raf-stall/clock-skew/delay-reset in the overlay; three regression tests fail on the old player. Not confirmed on the user's browser (U). Build id: dev-update stamps the shipped commit (--stamp) instead of the worker mirror's VCS commit; tested."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 player compares with performance.now only, fallback timer when animation frames are withheld, inflate-stall reconnect; regression tests fail on the old player
- [x] #2 dev update stamps the shipped commit as the page build id (not the worker mirror's)
<!-- AC:END -->
