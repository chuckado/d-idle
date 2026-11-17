---
id: TASK-620
title: 'videoctl: video controls out of the picture (toolbar below), caption tap to expand, fit rule'
status: Proposed
assignee:
  - credit-runner
created_date: '2026-11-17T15:01:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 310000
card: "videoctl"
owns:
  - "internal/watch/web/panels/video"
  - "tools/uishots"
  - "tools/uishots.sh"
  - "tools/videofeed"
  - "docs/video-feed.md"
  - "internal/watch/testdata/video_panel.js"
  - "internal/watch/testdata/video_pal_panel.js"
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
- [ ] #1 controls in a toolbar below the picture on all sizes; caption never covered; tap shows the full caption
- [ ] #2 fit rule: a control touching the caption or other text of the video card fails; must-fail defect page; long-caption fixture
<!-- AC:END -->
