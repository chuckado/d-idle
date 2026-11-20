---
id: TASK-678
title: 'viewerconns: one video transport per page, probe before streaming'
status: In Progress
assignee:
  - credit-runner
created_date: '2026-11-19T22:34:00Z'
labels:
  - watch
  - bug
dependencies: []
priority: medium
ordinal: 331000
card: "viewerconns"
result: "Not verified live. Headless Chrome on the builder against the uishots fixture: the page holds 2 long-lived requests (events + one video feed), no growth over 2 reloads. The panel now probes before opening a stream and aborts a stalled probe after 6 s; the old code held the MJPEG stream beside the pending probe (must-fail test video_conns.js). The cause of the stalled probe on the real host is not proven."
owns:
  - "internal/watch/web/panels/video"
  - "internal/watch/testdata/video_conns.js"
  - "tools/connprobe"
exclusive: []
contracts: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Size: M

Viewer NO SIGNAL: MJPEG plus a pending palette probe held connections beside the event stream. Done when the page holds <= 2 long-lived requests.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The video panel opens at most one video transport (palette feed or MJPEG) and probes before any stream
- [x] #2 Must-fail test counts open streams and the probe timeout
- [x] #3 connprobe harness counts the page's long-lived requests (2: events + video)
<!-- AC:END -->
