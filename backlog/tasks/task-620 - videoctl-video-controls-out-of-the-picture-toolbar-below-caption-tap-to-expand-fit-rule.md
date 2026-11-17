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
result: "The video card's buttons were the title-bar actions of a cockpit slot, painted over the top of the picture and wrapping across the quest caption on a phone. Now a one-row scrolling toolbar BELOW the picture (window.d2video.controls.add for new buttons, e.g. Labels), only Collapse stays in the title bar; the low-fps hint and the rate line are one line each; the caption keeps its corner and a tap shows the whole line. Fit rule: a control touching the caption or any other text of the video card fails; the viddefect-ctl page fails as it must at all sizes; default-live-pal and quest-running pass fit at all sizes; pictures viewed at 412x780, 360x640, 390x844 (checked via 360/412 and landscape), 844x390, 1366x768, 1914x930. quest-runlong fit pages were added after the last fit run (not run)."
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Why: TODO. Done when: TODO.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 controls in a toolbar below the picture on all sizes; caption never covered; tap shows the full caption
- [x] #2 fit rule: a control touching the caption or other text of the video card fails; must-fail defect page; long-caption fixture
<!-- AC:END -->
